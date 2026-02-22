import { Plane, MapPin, Clock, Luggage, QrCode, Download } from 'lucide-react';

export default function BoardingPass({ boardingPass, booking }) {
  const formatDate = (dateString) => {
    const date = new Date(dateString);
    return date.toLocaleDateString('en-US', { 
      weekday: 'short', 
      month: 'short', 
      day: 'numeric',
      year: 'numeric'
    });
  };

  const formatTime = (dateString) => {
    const date = new Date(dateString);
    return date.toLocaleTimeString('en-US', { 
      hour: '2-digit', 
      minute: '2-digit'
    });
  };

  return (
    <div className="min-h-screen bg-gradient-to-br from-green-50 via-emerald-50 to-teal-50 flex items-center justify-center p-4">
      <div className="max-w-2xl w-full">
        {/* Success Message */}
        <div className="text-center mb-8">
          <div className="inline-flex bg-green-100 p-4 rounded-full mb-4">
            <svg className="w-16 h-16 text-green-600" fill="none" stroke="currentColor" viewBox="0 0 24 24">
              <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M5 13l4 4L19 7" />
            </svg>
          </div>
          <h1 className="text-3xl font-bold text-gray-900 mb-2">
            Check-In Successful!
          </h1>
          <p className="text-gray-600">
            Your boarding pass is ready
          </p>
        </div>

        {/* Boarding Pass */}
        <div className="bg-white rounded-2xl shadow-2xl overflow-hidden">
          {/* Header */}
          <div className="bg-gradient-to-r from-blue-600 to-indigo-600 p-6 text-white">
            <div className="flex items-center justify-between">
              <div>
                <p className="text-sm opacity-90">SkyHigh Airlines</p>
                <h2 className="text-2xl font-bold">{boardingPass.flightCode}</h2>
              </div>
              <Plane className="w-10 h-10" />
            </div>
          </div>

          {/* Flight Route */}
          <div className="p-6 border-b border-gray-200">
            <div className="flex items-center justify-between">
              <div className="text-center">
                <p className="text-3xl font-bold text-gray-900">{boardingPass.source}</p>
                <p className="text-sm text-gray-500 mt-1">Departure</p>
              </div>
              
              <div className="flex flex-col items-center px-4">
                <Plane className="w-6 h-6 text-gray-400 rotate-90" />
                <div className="w-24 h-px bg-gray-300 my-2"></div>
              </div>
              
              <div className="text-center">
                <p className="text-3xl font-bold text-gray-900">{boardingPass.destination}</p>
                <p className="text-sm text-gray-500 mt-1">Arrival</p>
              </div>
            </div>
          </div>

          {/* Passenger Details */}
          <div className="p-6 bg-gray-50 border-b border-gray-200">
            <div className="grid grid-cols-2 gap-6">
              <div>
                <p className="text-xs text-gray-500 uppercase tracking-wide mb-1">Passenger Name</p>
                <p className="text-lg font-semibold text-gray-900">{boardingPass.passengerName}</p>
              </div>
              <div>
                <p className="text-xs text-gray-500 uppercase tracking-wide mb-1">PNR</p>
                <p className="text-lg font-mono font-semibold text-gray-900">{boardingPass.pnr}</p>
              </div>
            </div>
          </div>

          {/* Flight Details Grid */}
          <div className="p-6">
            <div className="grid grid-cols-3 gap-6">
              {/* Seat */}
              <div className="text-center">
                <p className="text-xs text-gray-500 uppercase tracking-wide mb-2">Seat</p>
                <div className="bg-blue-100 text-blue-900 text-3xl font-bold py-3 rounded-lg">
                  {boardingPass.seat}
                </div>
              </div>

              {/* Gate */}
              <div className="text-center">
                <p className="text-xs text-gray-500 uppercase tracking-wide mb-2">Gate</p>
                <div className="bg-green-100 text-green-900 text-3xl font-bold py-3 rounded-lg">
                  {boardingPass.gate}
                </div>
              </div>

              {/* Boarding Time */}
              <div className="text-center">
                <p className="text-xs text-gray-500 uppercase tracking-wide mb-2">Boarding</p>
                <div className="bg-orange-100 text-orange-900 text-xl font-bold py-3 rounded-lg flex items-center justify-center">
                  <Clock className="w-5 h-5 mr-2" />
                  {boardingPass.boardingTime}
                </div>
              </div>
            </div>

            {/* Departure Info */}
            <div className="mt-6 pt-6 border-t border-gray-200">
              <div className="grid grid-cols-2 gap-4 text-sm">
                <div>
                  <p className="text-gray-500 mb-1">Date</p>
                  <p className="font-semibold text-gray-900">{formatDate(boardingPass.departureTime)}</p>
                </div>
                <div>
                  <p className="text-gray-500 mb-1">Departure Time</p>
                  <p className="font-semibold text-gray-900">{formatTime(boardingPass.departureTime)}</p>
                </div>
              </div>
            </div>
          </div>

          {/* QR Code Placeholder */}
          <div className="p-6 bg-gray-50 border-t border-gray-200">
            <div className="flex justify-center">
              <div className="bg-white p-4 rounded-lg border-2 border-dashed border-gray-300">
                <QrCode className="w-32 h-32 text-gray-400" />
              </div>
            </div>
            <p className="text-center text-xs text-gray-500 mt-3">Scan at boarding gate</p>
          </div>

          {/* Barcode Simulation */}
          <div className="bg-white p-4">
            <div className="h-16 bg-gradient-to-r from-gray-800 via-gray-600 to-gray-800 rounded-lg flex items-center justify-center">
              <div className="flex gap-1">
                {[...Array(30)].map((_, i) => (
                  <div 
                    key={i} 
                    className="bg-white h-12" 
                    style={{ width: Math.random() > 0.5 ? '2px' : '4px' }}
                  />
                ))}
              </div>
            </div>
          </div>
        </div>

        {/* Actions */}
        <div className="mt-8 flex gap-4">
          <button
            onClick={() => window.location.href = '/'}
            className="flex-1 bg-white text-gray-700 font-semibold py-4 rounded-lg border-2 border-gray-300 hover:border-gray-400 transition-all"
          >
            Back to Home
          </button>
          <button
            onClick={() => window.print()}
            className="flex-1 bg-gradient-to-r from-blue-600 to-indigo-600 text-white font-semibold py-4 rounded-lg hover:from-blue-700 hover:to-indigo-700 transition-all flex items-center justify-center gap-2"
          >
            <Download className="w-5 h-5" />
            Print / Save
          </button>
        </div>

        {/* Important Notes */}
        <div className="mt-6 bg-yellow-50 border-2 border-yellow-200 rounded-lg p-4">
          <p className="text-sm text-yellow-800">
            ⚠️ <strong>Important:</strong> Please arrive at the gate at least 30 minutes before boarding time.
          </p>
        </div>
      </div>
    </div>
  );
}
