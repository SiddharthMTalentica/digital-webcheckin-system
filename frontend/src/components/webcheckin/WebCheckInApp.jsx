import React, { useState, useEffect } from 'react';
import { useSearchParams, useNavigate } from 'react-router-dom';
import PNRLookup from './PNRLookup';
import SeatMap from '../SeatMap';
import PaymentModal from './PaymentModal';
import BoardingPass from './BoardingPass';

export default function WebCheckInApp() {
  const [searchParams] = useSearchParams();
  const navigate = useNavigate();
  const urlPnr = searchParams.get('pnr') || '';
  const urlLastName = searchParams.get('lastName') || '';

  const [step, setStep] = useState('lookup'); // lookup, seats, payment, complete
  const [booking, setBooking] = useState(null);
  const [seats, setSeats] = useState([]);
  const [selectedSeat, setSelectedSeat] = useState(null);
  const [heldSeat, setHeldSeat] = useState(null);
  const [timeLeft, setTimeLeft] = useState(120); // 120 seconds
  const [baggageWeight, setBaggageWeight] = useState('');
  const [checkInId, setCheckInId] = useState(null);
  const [feeDetails, setFeeDetails] = useState(null);
  const [boardingPass, setBoardingPass] = useState(null);
  const [error, setError] = useState(null);
  const [loading, setLoading] = useState(false);

  // Transform webcheckin seat data to match SeatMap format
  const transformSeat = (seat) => {
    const seatNo = seat.seat_no || seat.seatNo || '';
    const match = seatNo.match(/^(\d+)([A-Z])$/);
    
    let mappedStatus = seat.is_booked ? 'CONFIRMED' : (seat.checkInId ? 'HELD' : 'AVAILABLE');
    
    // If this seat was booked by the current user initially, make it available for them to keep
    if (booking?.initialSeatNo === seatNo) {
        mappedStatus = 'AVAILABLE';
    }

    return {
      ...seat,
      seat_no: seatNo,
      row_num: seat.row_num || (match ? parseInt(match[1], 10) : 0),
      col_num: seat.col_num || (match ? match[2] : ''),
      category: seat.category || (seat.isPremium ? 'PREMIUM' : 'STANDARD'),
      status: mappedStatus,
      is_booked: seat.is_booked || false,
    };
  };

  // Fetch seats when booking is set
  useEffect(() => {
    if (booking && step === 'seats') {
      const fetchSeats = async () => {
        setLoading(true);
        try {
          const response = await fetch(
            `http://localhost:8081/api/webcheckin/${booking.pnr}/seats?flightId=${booking.flight.id}`
          );
          const data = await response.json();
          if (response.ok) {
            const rawSeats = data.seats || [];
            const transformedSeats = rawSeats.map(transformSeat);
            setSeats(transformedSeats);
            
            // Auto-select initial seat if no seat is selected yet
            if (booking.initialSeatNo && !selectedSeat) {
              const prevSeat = transformedSeats.find(s => s.seat_no === booking.initialSeatNo);
              if (prevSeat) {
                setSelectedSeat(prevSeat);
                // Also set heldSeat so it shows as properly 'Reserved'/'Your Seat' on the SeatMap like other held seats
                setHeldSeat(prevSeat);
                setTimeLeft(86400); // arbitrarily large time since it's already theirs
              }
            }
          } else {
            setError('Failed to load seats');
          }
        } catch (err) {
          setError('Failed to load seats');
        } finally {
          setLoading(false);
        }
      };
      fetchSeats();
    }
  }, [booking, step]);

  // Handle successful PNR lookup
  const handleLookupSuccess = (bookingData) => {
    setBooking(bookingData);
    setStep('seats');
  };

  // Handle seat selection
  const handleSeatSelect = async (seat) => {
    if (seat.status !== 'AVAILABLE' || seat.checkInId) {
      setError('Seat not available');
      return;
    }

    setError(null);

    try {
      const response = await fetch(
        `http://localhost:8081/api/webcheckin/${booking.pnr}/hold-seat`,
        {
          method: 'POST',
          headers: { 'Content-Type': 'application/json' },
          body: JSON.stringify({
            flightId: booking.flight.id,
            seatNo: seat.seat_no,
          }),
        }
      );

      const data = await response.json();

      if (!response.ok) {
        throw new Error(data.error || 'Failed to hold seat');
      }

      setSelectedSeat(seat);
      setHeldSeat(seat);
      setTimeLeft(data.holdDuration);
    } catch (err) {
      setError(err.message);
    }
  };

  // Timer countdown
  React.useEffect(() => {
    if (step === 'seats' && heldSeat && timeLeft > 0) {
      const timer = setInterval(() => {
        setTimeLeft((prev) => {
          if (prev <= 1) {
            setHeldSeat(null);
            setSelectedSeat(null);
            setError('Seat hold expired. Please select again.');
            return 0;
          }
          return prev - 1;
        });
      }, 1000);

      return () => clearInterval(timer);
    }
  }, [step, heldSeat, timeLeft]);

  // Handle baggage submission
  const handleBaggageSubmit = async () => {
    if (!baggageWeight || baggageWeight < 0) {
      setError('Please enter a valid baggage weight');
      return;
    }

    setError(null);

    try {
      const response = await fetch(
        `http://localhost:8081/api/webcheckin/${booking.pnr}/complete`,
        {
          method: 'POST',
          headers: { 'Content-Type': 'application/json' },
          body: JSON.stringify({
            seatNo: selectedSeat.seat_no,
            baggageWeight: parseFloat(baggageWeight),
          }),
        }
      );

      const data = await response.json();

      if (response.status === 402) {
        // Payment required
        setCheckInId(data.checkInId);
        setFeeDetails(data.details);
        setStep('payment');
        return;
      }

      if (!response.ok) {
        throw new Error(data.error || 'Check-in failed');
      }

      // Success - go to boarding pass
      setBoardingPass(data.boardingPass);
      setStep('complete');
      setTimeLeft(0);
    } catch (err) {
      setError(err.message);
    }
  };

  // Handle payment success
  const handlePaymentSuccess = (boardingPassData) => {
    setBoardingPass(boardingPassData);
    setStep('complete');
    setTimeLeft(0);
  };

  // Render based on step
  if (step === 'lookup') {
    return <PNRLookup onLookupSuccess={handleLookupSuccess} initialPnr={urlPnr} initialLastName={urlLastName} />;
  }

  if (step === 'complete') {
    return <BoardingPass boardingPass={boardingPass} booking={booking} />;
  }

  // Seats + Baggage step (combined)
  return (
    <div className="min-h-screen bg-gradient-to-br from-blue-50 via-indigo-50 to-purple-50 p-4">
      <div className="max-w-6xl mx-auto">
        {/* Header */}
        <div className="bg-white rounded-2xl shadow-lg p-6 mb-6">
          <div className="flex items-center justify-between">
            <div>
              <h1 className="text-2xl font-bold text-gray-900">Web Check-In</h1>
              <p className="text-gray-600 mt-1">
                {booking.flight.source} → {booking.flight.destination} · Flight {booking.flight.code}
              </p>
            </div>
            <div className="text-right">
              <p className="text-sm text-gray-500">Passenger</p>
              <p className="font-semibold text-gray-900">
                {booking.passenger.firstName} {booking.passenger.lastName}
              </p>
              <p className="text-xs text-gray-500 font-mono">PNR: {booking.pnr}</p>
            </div>
          </div>
        </div>

        {/* Progress Steps */}
        <div className="bg-white rounded-2xl shadow-lg p-6 mb-6">
          <div className="flex items-center justify-between">
            <div className="flex items-center">
              <div className="bg-blue-600 text-white w-8 h-8 rounded-full flex items-center justify-center font-semibold">
                1
              </div>
              <span className="ml-3 font-semibold text-gray-900">Select Seat</span>
            </div>
            <div className="flex-1 h-px bg-gray-300 mx-4"></div>
            <div className="flex items-center">
              <div className={`${selectedSeat ? 'bg-blue-600 text-white' : 'bg-gray-200 text-gray-500'} w-8 h-8 rounded-full flex items-center justify-center font-semibold`}>
                2
              </div>
              <span className={`ml-3 font-semibold ${selectedSeat ? 'text-gray-900' : 'text-gray-400'}`}>
                Baggage Details
              </span>
            </div>
            <div className="flex-1 h-px bg-gray-300 mx-4"></div>
            <div className="flex items-center">
              <div className="bg-gray-200 text-gray-500 w-8 h-8 rounded-full flex items-center justify-center font-semibold">
                3
              </div>
              <span className="ml-3 font-semibold text-gray-400">Complete</span>
            </div>
          </div>
        </div>

        {/* Main Content */}
        <div className="grid lg:grid-cols-3 gap-6">
          {/* Seat Map */}
          <div className="lg:col-span-2">
            {loading ? (
              <div className="bg-white rounded-2xl shadow-lg p-12 text-center">
                <div className="animate-spin w-12 h-12 border-4 border-blue-600 border-t-transparent rounded-full mx-auto mb-4"></div>
                <p className="text-gray-600">Loading seats...</p>
              </div>
            ) : (
              <SeatMap
                seats={seats}
                onSelectSeat={handleSeatSelect}
                selectedSeat={selectedSeat}
                heldSeat={heldSeat}
              />
            )}
          </div>

          {/* Baggage Form */}
          <div>
            <div className="bg-white rounded-2xl shadow-lg p-6">
              <h2 className="text-xl font-bold text-gray-900 mb-4">Baggage Details</h2>
              
              {!selectedSeat ? (
                <div className="text-center py-8">
                  <p className="text-gray-500">Select a seat to continue</p>
                </div>
              ) : (
                <div className="space-y-6">
                  {/* Selected Seat Info */}
                  <div className="bg-blue-50 rounded-lg p-4">
                    <p className="text-sm text-gray-600 mb-1">Selected Seat</p>
                    <p className="text-3xl font-bold text-blue-600">{selectedSeat.seat_no}</p>
                  </div>

                  {/* Baggage Weight Input */}
                  <div>
                    <label className="block text-sm font-semibold text-gray-700 mb-2">
                      Baggage Weight (kg)
                    </label>
                    <input
                      type="number"
                      min="0"
                      max="50"
                      step="0.1"
                      value={baggageWeight}
                      onChange={(e) => setBaggageWeight(e.target.value)}
                      placeholder="25.0"
                      className="w-full px-4 py-3 rounded-lg border-2 border-gray-200 focus:border-blue-500 focus:ring-2 focus:ring-blue-200 outline-none transition-all"
                    />
                    <p className="text-xs text-gray-500 mt-1">
                      Free allowance: 25 kg. Excess charged at $10/kg
                    </p>
                  </div>

                  {/* Error */}
                  {error && (
                    <div className="bg-red-50 border-2 border-red-200 rounded-lg p-3 text-sm text-red-800">
                      {error}
                    </div>
                  )}

                  {/* Submit Button */}
                  <button
                    onClick={handleBaggageSubmit}
                    disabled={!baggageWeight}
                    className="w-full bg-gradient-to-r from-blue-600 to-indigo-600 text-white font-semibold py-4 rounded-lg hover:from-blue-700 hover:to-indigo-700 transition-all disabled:opacity-50 disabled:cursor-not-allowed shadow-lg hover:shadow-xl"
                  >
                    Complete Check-In
                  </button>
                </div>
              )}
            </div>
          </div>
        </div>

        {/* Payment Modal */}
        {step === 'payment' && (
          <PaymentModal
            checkInId={checkInId}
            feeDetails={feeDetails}
            booking={booking}
            seatNo={selectedSeat.seatNo}
            onPaymentSuccess={handlePaymentSuccess}
          />
        )}
      </div>
    </div>
  );
}
