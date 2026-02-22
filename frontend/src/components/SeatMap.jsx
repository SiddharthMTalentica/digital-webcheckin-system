import React from 'react';

const SeatMap = ({ seats, onSelectSeat, selectedSeat, heldSeat }) => {
  // Group seats by row
  const rows = seats.reduce((acc, seat) => {
    const row = seat.row_num;
    if (!acc[row]) acc[row] = [];
    acc[row].push(seat);
    return acc;
  }, {});

  // Sort columns A,B,C...
  Object.keys(rows).forEach(row => {
    rows[row].sort((a, b) => a.col_num.localeCompare(b.col_num));
  });

  const getSeatColor = (seat) => {
    // CRITICAL FIX: Check status field first (which comes from backend)
    if (seat.status === 'CONFIRMED') {
      return 'bg-red-500 text-white cursor-not-allowed border-red-600';
    }
    
    // If held by others (status is HELD but not by me)
    if (seat.status === 'HELD' && heldSeat?.seat_no !== seat.seat_no) {
      return 'bg-amber-500 text-white cursor-not-allowed border-amber-600';
    }
    
    // If held by me
    if (heldSeat?.seat_no === seat.seat_no) {
      return 'bg-emerald-500 text-white border-emerald-600 ring-2 ring-emerald-300 shadow-lg';
    }
    
    // Selected UI state (before hold)
    if (selectedSeat?.seat_no === seat.seat_no) {
      return 'bg-blue-500 text-white border-blue-600';
    }
    
    // Available - differentiate between premium and standard
    if (seat.category === 'PREMIUM') {
      return 'bg-blue-50 text-gray-700 border-blue-200 hover:bg-blue-100 hover:border-blue-400 cursor-pointer';
    }
    
    return 'bg-slate-50 text-gray-700 border-slate-200 hover:bg-slate-100 hover:border-slate-400 cursor-pointer';
  };

  const getSeatIcon = (seat) => {
    if (seat.status === 'CONFIRMED') {
      return (
        <svg className="w-4 h-4 absolute top-0.5 right-0.5" fill="currentColor" viewBox="0 0 20 20">
          <path fillRule="evenodd" d="M5 9V7a5 5 0 0110 0v2a2 2 0 012 2v5a2 2 0 01-2 2H5a2 2 0 01-2-2v-5a2 2 0 012-2zm8-2v2H7V7a3 3 0 016 0z" clipRule="evenodd" />
        </svg>
      );
    }
    if (seat.status === 'HELD') {
      return (
        <svg className="w-4 h-4 absolute top-0.5 right-0.5" fill="currentColor" viewBox="0 0 20 20">
          <path fillRule="evenodd" d="M10 18a8 8 0 100-16 8 8 0 000 16zm1-12a1 1 0 10-2 0v4a1 1 0 00.293.707l2.828 2.829a1 1 0 101.415-1.415L11 9.586V6z" clipRule="evenodd" />
        </svg>
      );
    }
    if (seat.category === 'PREMIUM') {
      return (
        <svg className="w-4 h-4 absolute top-0.5 right-0.5 text-blue-400" fill="currentColor" viewBox="0 0 20 20">
          <path d="M9.049 2.927c.3-.921 1.603-.921 1.902 0l1.07 3.292a1 1 0 00.95.69h3.462c.969 0 1.371 1.24.588 1.81l-2.8 2.034a1 1 0 00-.364 1.118l1.07 3.292c.3.921-.755 1.688-1.54 1.118l-2.8-2.034a1 1 0 00-1.175 0l-2.8 2.034c-.784.57-1.838-.197-1.539-1.118l1.07-3.292a1 1 0 00-.364-1.118L2.98 8.72c-.783-.57-.38-1.81.588-1.81h3.461a1 1 0 00.951-.69l1.07-3.292z" />
        </svg>
      );
    }
    return null;
  };

  const isDisabled = (seat) => {
    return seat.status === 'CONFIRMED' || (seat.status === 'HELD' && heldSeat?.seat_no !== seat.seat_no);
  };

  return (
    <div className="bg-white rounded-xl shadow-md p-8">
      <div className="flex flex-col items-center space-y-6">
        <div className="text-center">
          <h3 className="text-2xl font-bold text-gray-900 mb-2">Select Your Seat</h3>
          <p className="text-gray-600">Choose your preferred seat for this flight</p>
        </div>
        
        {/* Cockpit */}
        <div className="relative">
          <div className="w-24 h-20 bg-gradient-to-b from-blue-900 to-blue-700 rounded-t-full flex items-end justify-center pb-2">
            <div className="text-white text-xs font-semibold">COCKPIT</div>
          </div>
        </div>

        {/* Seat Grid */}
        <div className="space-y-3">
          {Object.keys(rows).sort((a, b) => parseInt(a) - parseInt(b)).map((rowNum) => (
            <div key={rowNum} className="flex items-center gap-3">
              <div className="w-8 text-center">
                <span className="text-sm font-bold text-gray-500">{rowNum}</span>
              </div>
              
              <div className="flex gap-2">
                {rows[rowNum].map((seat, index) => (
                  <React.Fragment key={seat.id}>
                    <button
                      disabled={isDisabled(seat)}
                      onClick={() => !isDisabled(seat) && onSelectSeat(seat)}
                      className={`
                        relative w-12 h-12 rounded-lg border-2 font-bold text-sm
                        transition-all duration-200 transform
                        ${getSeatColor(seat)}
                        ${!isDisabled(seat) ? 'hover:scale-110 active:scale-95' : ''}
                        ${heldSeat?.seat_no === seat.seat_no ? 'scale-110' : ''}
                      `}
                      title={`Seat ${seat.seat_no} - ${seat.status} - ${seat.category}`}
                    >
                      {seat.col_num}
                      {getSeatIcon(seat)}
                    </button>
                    {/* Aisle gap after 3rd seat (index 2) */}
                    {index === 2 && <div className="w-6" />}
                  </React.Fragment>
                ))}
              </div>
              
              <div className="w-8 text-center">
                <span className="text-sm font-bold text-gray-500">{rowNum}</span>
              </div>
            </div>
          ))}
        </div>

        {/* Legend */}
        <div className="w-full pt-6 border-t border-gray-200">
          <h4 className="text-sm font-semibold text-gray-700 mb-3">Legend</h4>
          <div className="grid grid-cols-2 md:grid-cols-4 gap-3">
            <div className="flex items-center space-x-2">
              <div className="w-8 h-8 bg-slate-50 border-2 border-slate-200 rounded"></div>
              <span className="text-sm text-gray-600">Available</span>
            </div>
            <div className="flex items-center space-x-2">
              <div className="w-8 h-8 bg-blue-50 border-2 border-blue-200 rounded relative">
                <svg className="w-3 h-3 absolute top-0.5 right-0.5 text-blue-400" fill="currentColor" viewBox="0 0 20 20">
                  <path d="M9.049 2.927c.3-.921 1.603-.921 1.902 0l1.07 3.292a1 1 0 00.95.69h3.462c.969 0 1.371 1.24.588 1.81l-2.8 2.034a1 1 0 00-.364 1.118l1.07 3.292c.3.921-.755 1.688-1.54 1.118l-2.8-2.034a1 1 0 00-1.175 0l-2.8 2.034c-.784.57-1.838-.197-1.539-1.118l1.07-3.292a1 1 0 00-.364-1.118L2.98 8.72c-.783-.57-.38-1.81.588-1.81h3.461a1 1 0 00.951-.69l1.07-3.292z" />
                </svg>
              </div>
              <span className="text-sm text-gray-600">Premium</span>
            </div>
            <div className="flex items-center space-x-2">
              <div className="w-8 h-8 bg-emerald-500 border-2 border-emerald-600 rounded"></div>
              <span className="text-sm text-gray-600">Your Seat</span>
            </div>
            <div className="flex items-center space-x-2">
              <div className="w-8 h-8 bg-amber-500 border-2 border-amber-600 rounded"></div>
              <span className="text-sm text-gray-600">Reserved</span>
            </div>
            <div className="flex items-center space-x-2">
              <div className="w-8 h-8 bg-red-500 border-2 border-red-600 rounded"></div>
              <span className="text-sm text-gray-600">Occupied</span>
            </div>
          </div>
        </div>
      </div>
    </div>
  );
};

export default SeatMap;
