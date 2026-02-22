import { useState, useEffect } from 'react';
import { useNavigate } from 'react-router-dom';
import { getFlights, getSeatMap, holdSeat, confirmCheckIn } from './lib/api';
import SeatMap from './components/SeatMap';

// Simple UUID generator for demo
const generateUserId = () => 'user-' + Math.random().toString(36).substr(2, 9);

function App() {
  const navigate = useNavigate();
  const [step, setStep] = useState(1); // 1: Flight, 2: Seat, 3: Passenger, 4: Success
  const [flights, setFlights] = useState([]);
  const [filteredFlights, setFilteredFlights] = useState([]);
  const [selectedDate, setSelectedDate] = useState(new Date().toISOString().split('T')[0]);
  const [selectedFlight, setSelectedFlight] = useState(null);
  const [seats, setSeats] = useState([]);
  const [selectedSeat, setSelectedSeat] = useState(null);
  const [heldSeat, setHeldSeat] = useState(null);
  const [timeLeft, setTimeLeft] = useState(0);
  const [userId] = useState(generateUserId());
  const [bookingResult, setBookingResult] = useState(null);
  const [copied, setCopied] = useState(null);
  
  const [formData, setFormData] = useState({
    firstName: '',
    lastName: '',
    passport: '',
    baggageWeight: 0
  });

  const [loading, setLoading] = useState(false);
  const [error, setError] = useState(null);

  const copyToClipboard = (text, label) => {
    navigator.clipboard.writeText(text).then(() => {
      setCopied(label);
      setTimeout(() => setCopied(null), 2000);
    });
  };

  // Load flights
  useEffect(() => {
    getFlights().then(data => {
      setFlights(data);
      filterFlightsByDate(data, selectedDate);
    }).catch(err => setError(err.message));
  }, []);

  // Filter flights by date
  const filterFlightsByDate = (flightList, date) => {
    const filtered = flightList.filter(f => {
      const flightDate = new Date(f.departure_time).toISOString().split('T')[0];
      return flightDate === date;
    });
    setFilteredFlights(filtered);
  };

  const handleDateChange = (e) => {
    const newDate = e.target.value;
    setSelectedDate(newDate);
    filterFlightsByDate(flights, newDate);
  };

  // Timer Logic
  useEffect(() => {
    if (step === 3 && heldSeat && timeLeft > 0) {
      const interval = setInterval(() => {
        setTimeLeft(prev => prev - 1);
      }, 1000);
      return () => clearInterval(interval);
    } else if (step === 3 && heldSeat && timeLeft === 0) {
      setHeldSeat(null);
      setError("Session expired. Please select seat again.");
      setStep(2);
      refreshSeats(selectedFlight.id);
    }
  }, [timeLeft, heldSeat, step]);

  const refreshSeats = async (flightId) => {
      setLoading(true);
      try {
        const data = await getSeatMap(flightId);
        setSeats(data.seats);
      } catch (err) {
          setError("Failed to load seat map");
      } finally {
          setLoading(false);
      }
  };

  const handleFlightSelect = async (flight) => {
    setSelectedFlight(flight);
    setStep(2);
    await refreshSeats(flight.id);
  };

  const handleSeatSelect = async (seat) => {
      setSelectedSeat(seat);
      setLoading(true);
      setError(null);
      try {
          await holdSeat(selectedFlight.id, seat.seat_no, userId);
          setHeldSeat(seat);
          setTimeLeft(45);
          setStep(3);
      } catch (err) {
          setError(err.response?.data || "Failed to hold seat");
          await refreshSeats(selectedFlight.id);
      } finally {
          setLoading(false);
      }
  };

  const handleSkipSeat = () => {
    setSelectedSeat(null);
    setHeldSeat(null);
    setStep(3);
    setTimeLeft(0);
  };

  const handleConfirm = async (e) => {
      e.preventDefault();
      setLoading(true);
      setError(null);
      
      try {
          const res = await confirmCheckIn({
              flight_id: selectedFlight.id,
              seat_no: heldSeat ? heldSeat.seat_no : "",
              user_id: userId,
              first_name: formData.firstName,
              last_name: formData.lastName,
              passport: formData.passport,
              baggage_weight: parseFloat(formData.baggageWeight)
          });
          // Show success card on screen instead of alert
          setBookingResult({
            pnr: res.pnr,
            bookingReference: res.booking_reference,
            flight: selectedFlight,
            seatNo: heldSeat ? heldSeat.seat_no : "Unassigned",
            passenger: `${formData.firstName} ${formData.lastName}`,
            lastName: formData.lastName,
          });
          setHeldSeat(null);
          setTimeLeft(0);
          setStep(4);
      } catch (err) {
          if (err.response?.status === 402) {
               if(window.confirm(`Overweight baggage! Pay fee for ${formData.baggageWeight}kg?`)) {
                    alert("Payment Simulated: Successful!");
                    setError("Baggage limits enforced for free check-in. Please reduce weight or contact staff.");
               } else {
                  setError("Payment declined.");
               }
          } else {
              setError(err.response?.data || "Check-in failed");
          }
      } finally {
          setLoading(false);
      }
  };

  const getTimerColor = () => {
    if (timeLeft > 30) return 'text-emerald-600';
    if (timeLeft > 15) return 'text-amber-600';
    return 'text-red-600';
  };

  return (
    <div className="min-h-screen bg-gradient-to-br from-slate-50 to-slate-100">
      {/* Header */}
      <header className="bg-gradient-to-r from-blue-900 to-blue-700 text-white shadow-lg">
        <div className="max-w-7xl mx-auto px-4 sm:px-6 lg:px-8 py-6">
          <div className="flex items-center justify-between">
            <div className="flex items-center space-x-3 cursor-pointer" onClick={() => navigate('/')}>
              <svg className="w-10 h-10" fill="currentColor" viewBox="0 0 20 20">
                <path d="M10.894 2.553a1 1 0 00-1.788 0l-7 14a1 1 0 001.169 1.409l5-1.429A1 1 0 009 15.571V11a1 1 0 112 0v4.571a1 1 0 00.725.962l5 1.428a1 1 0 001.17-1.408l-7-14z" />
              </svg>
              <div>
                <h1 className="text-3xl font-bold tracking-tight">SkyHigh Airlines</h1>
                <p className="text-blue-200 text-sm">Flight Booking System</p>
              </div>
            </div>
            <div className="flex items-center space-x-3">
              <button
                onClick={() => navigate('/')}
                className="px-4 py-2 bg-white/10 hover:bg-white/20 rounded-lg transition-colors duration-200 flex items-center space-x-2"
                title="Home"
              >
                <svg className="w-5 h-5" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                  <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M3 12l2-2m0 0l7-7 7 7M5 10v10a1 1 0 001 1h3m10-11l2 2m-2-2v10a1 1 0 01-1 1h-3m-6 0a1 1 0 001-1v-4a1 1 0 011-1h2a1 1 0 011 1v4a1 1 0 001 1m-6 0h6" />
                </svg>
                <span>Home</span>
              </button>
              {step > 1 && step < 4 && (
                <button
                  onClick={() => {
                    setStep(1);
                    setError(null);
                    setHeldSeat(null);
                    setTimeLeft(0);
                  }}
                  className="px-4 py-2 bg-white/10 hover:bg-white/20 rounded-lg transition-colors duration-200 flex items-center space-x-2"
                >
                  <svg className="w-5 h-5" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                    <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M10 19l-7-7m0 0l7-7m-7 7h18" />
                  </svg>
                  <span>Back to Flights</span>
                </button>
              )}
            </div>
          </div>
        </div>
      </header>

      {/* Timer Banner */}
      {timeLeft > 0 && (
        <div className="bg-gradient-to-r from-amber-500 to-orange-500 text-white shadow-md">
          <div className="max-w-7xl mx-auto px-4 py-3">
            <div className="flex items-center justify-between">
              <div className="flex items-center space-x-3">
                <svg className="w-6 h-6 animate-pulse" fill="currentColor" viewBox="0 0 20 20">
                  <path fillRule="evenodd" d="M10 18a8 8 0 100-16 8 8 0 000 16zm1-12a1 1 0 10-2 0v4a1 1 0 00.293.707l2.828 2.829a1 1 0 101.415-1.415L11 9.586V6z" clipRule="evenodd" />
                </svg>
                <span className="font-semibold">Seat Reserved</span>
              </div>
              <div className={`text-2xl font-bold ${getTimerColor()}`}>
                {Math.floor(timeLeft / 60)}:{(timeLeft % 60).toString().padStart(2, '0')}
              </div>
            </div>
            <div className="mt-2 bg-white/20 rounded-full h-2 overflow-hidden">
              <div 
                className="h-full bg-white transition-all duration-1000"
                style={{ width: `${(timeLeft / 45) * 100}%` }}
              />
            </div>
          </div>
        </div>
      )}

      {/* Error Alert */}
      {error && (
        <div className="max-w-7xl mx-auto px-4 mt-4">
          <div className="bg-red-50 border-l-4 border-red-500 p-4 rounded-r-lg">
            <div className="flex items-center">
              <svg className="w-5 h-5 text-red-500 mr-3" fill="currentColor" viewBox="0 0 20 20">
                <path fillRule="evenodd" d="M10 18a8 8 0 100-16 8 8 0 000 16zM8.707 7.293a1 1 0 00-1.414 1.414L8.586 10l-1.293 1.293a1 1 0 101.414 1.414L10 11.414l1.293 1.293a1 1 0 001.414-1.414L11.414 10l1.293-1.293a1 1 0 00-1.414-1.414L10 8.586 8.707 7.293z" clipRule="evenodd" />
              </svg>
              <p className="text-red-700 font-medium">{error}</p>
            </div>
          </div>
        </div>
      )}

      {/* Main Content */}
      <main className="max-w-7xl mx-auto px-4 sm:px-6 lg:px-8 py-8">
        
        {/* Step 1: Flight Selection */}
        {step === 1 && (
          <div className="space-y-6">
            {/* Date Selector */}
            <div className="bg-white rounded-xl shadow-md p-6">
              <label className="block text-sm font-semibold text-gray-700 mb-2">
                Select Travel Date
              </label>
              <input
                type="date"
                value={selectedDate}
                onChange={handleDateChange}
                className="w-full px-4 py-3 border-2 border-gray-200 rounded-lg focus:ring-2 focus:ring-blue-500 focus:border-transparent transition-all"
              />
            </div>

            {/* Flights List */}
            <div>
              <h2 className="text-2xl font-bold text-gray-900 mb-4">Available Flights</h2>
              {filteredFlights.length === 0 ? (
                <div className="bg-white rounded-xl shadow-md p-12 text-center">
                  <svg className="w-16 h-16 text-gray-400 mx-auto mb-4" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                    <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M20 13V6a2 2 0 00-2-2H6a2 2 0 00-2 2v7m16 0v5a2 2 0 01-2 2H6a2 2 0 01-2-2v-5m16 0h-2.586a1 1 0 00-.707.293l-2.414 2.414a1 1 0 01-.707.293h-3.172a1 1 0 01-.707-.293l-2.414-2.414A1 1 0 006.586 13H4" />
                  </svg>
                  <p className="text-gray-500 text-lg">No flights available for this date</p>
                </div>
              ) : (
                <div className="grid gap-4">
                  {filteredFlights.map(flight => (
                    <button
                      key={flight.id}
                      onClick={() => handleFlightSelect(flight)}
                      className="group bg-white rounded-xl shadow-md hover:shadow-xl transition-all duration-300 p-6 text-left border-2 border-transparent hover:border-blue-500"
                    >
                      <div className="flex items-center justify-between">
                        <div className="flex-1">
                          <div className="flex items-center space-x-4 mb-3">
                            <span className="text-3xl font-bold text-blue-600 group-hover:text-blue-700">
                              {flight.code}
                            </span>
                            <span className="px-3 py-1 bg-blue-100 text-blue-800 text-sm font-semibold rounded-full">
                              {flight.plane_type}
                            </span>
                          </div>
                          <div className="flex items-center space-x-4 text-gray-600">
                            <div className="flex items-center space-x-2">
                              <svg className="w-5 h-5" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                                <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M17.657 16.657L13.414 20.9a1.998 1.998 0 01-2.827 0l-4.244-4.243a8 8 0 1111.314 0z" />
                                <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M15 11a3 3 0 11-6 0 3 3 0 016 0z" />
                              </svg>
                              <span className="font-medium">{flight.source}</span>
                            </div>
                            <svg className="w-6 h-6 text-gray-400" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                              <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M17 8l4 4m0 0l-4 4m4-4H3" />
                            </svg>
                            <div className="flex items-center space-x-2">
                              <svg className="w-5 h-5" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                                <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M17.657 16.657L13.414 20.9a1.998 1.998 0 01-2.827 0l-4.244-4.243a8 8 0 1111.314 0z" />
                                <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M15 11a3 3 0 11-6 0 3 3 0 016 0z" />
                              </svg>
                              <span className="font-medium">{flight.destination}</span>
                            </div>
                          </div>
                        </div>
                        <div className="text-right space-y-2">
                          <div className="text-sm text-gray-500">Departure</div>
                          <div className="text-lg font-semibold text-gray-900">
                            {new Date(flight.departure_time).toLocaleTimeString('en-US', { 
                              hour: '2-digit', 
                              minute: '2-digit' 
                            })}
                          </div>
                          <div className="text-sm text-gray-500">
                            {new Date(flight.departure_time).toLocaleDateString('en-US', { 
                              month: 'short', 
                              day: 'numeric' 
                            })}
                          </div>
                        </div>
                      </div>
                    </button>
                  ))}
                </div>
              )}
            </div>
          </div>
        )}

        {/* Step 2: Seat Selection */}
        {step === 2 && selectedFlight && (
          <div className="space-y-6">
            <div className="bg-white rounded-xl shadow-md p-6 flex flex-col md:flex-row md:items-center justify-between gap-4">
              <div>
                <h2 className="text-2xl font-bold text-gray-900 mb-2">
                  Flight {selectedFlight.code}
                </h2>
                <p className="text-gray-600">
                  {selectedFlight.source} → {selectedFlight.destination}
                </p>
              </div>
              <button
                onClick={handleSkipSeat}
                className="px-6 py-2 bg-gray-100 hover:bg-gray-200 text-gray-700 font-semibold rounded-lg transition-colors border border-gray-300"
              >
                Skip Seat Selection
              </button>
            </div>
            {loading ? (
              <div className="bg-white rounded-xl shadow-md p-12 text-center">
                <div className="animate-spin rounded-full h-16 w-16 border-b-2 border-blue-600 mx-auto"></div>
                <p className="mt-4 text-gray-600">Loading seat map...</p>
              </div>
            ) : (
              <SeatMap 
                seats={seats} 
                heldSeat={heldSeat}
                selectedSeat={selectedSeat}
                onSelectSeat={handleSeatSelect} 
              />
            )}
          </div>
        )}

        {/* Step 3: Passenger Details */}
        {step === 3 && (
          <div className="max-w-2xl mx-auto">
            <div className="bg-white rounded-xl shadow-md p-8">
              <h2 className="text-2xl font-bold text-gray-900 mb-6">Passenger Details</h2>
              
              <div className="mb-6 p-4 bg-blue-50 rounded-lg border border-blue-200">
                <div className="flex items-center justify-between">
                  <span className="text-sm font-medium text-gray-700">Selected Seat</span>
                  <span className={`text-2xl font-bold ${heldSeat ? 'text-blue-600' : 'text-gray-500'}`}>
                    {heldSeat ? heldSeat.seat_no : "Unassigned (Select at Check-In)"}
                  </span>
                </div>
              </div>

              <form onSubmit={handleConfirm} className="space-y-6">
                <div className="grid grid-cols-2 gap-4">
                  <div className="relative">
                    <input
                      id="firstName"
                      type="text"
                      required
                      className="peer w-full px-4 py-3 border-2 border-gray-200 rounded-lg focus:ring-2 focus:ring-blue-500 focus:border-transparent transition-all placeholder-transparent"
                      placeholder="First Name"
                      value={formData.firstName}
                      onChange={e => setFormData({...formData, firstName: e.target.value})}
                    />
                    <label htmlFor="firstName" className="pointer-events-none absolute left-4 -top-2.5 bg-white px-2 text-sm font-medium text-gray-600 transition-all peer-placeholder-shown:top-3 peer-placeholder-shown:text-base peer-placeholder-shown:text-gray-400 peer-focus:-top-2.5 peer-focus:text-sm peer-focus:text-blue-600">
                      First Name
                    </label>
                  </div>
                  <div className="relative">
                    <input
                      id="lastName"
                      type="text"
                      required
                      className="peer w-full px-4 py-3 border-2 border-gray-200 rounded-lg focus:ring-2 focus:ring-blue-500 focus:border-transparent transition-all placeholder-transparent"
                      placeholder="Last Name"
                      value={formData.lastName}
                      onChange={e => setFormData({...formData, lastName: e.target.value})}
                    />
                    <label htmlFor="lastName" className="pointer-events-none absolute left-4 -top-2.5 bg-white px-2 text-sm font-medium text-gray-600 transition-all peer-placeholder-shown:top-3 peer-placeholder-shown:text-base peer-placeholder-shown:text-gray-400 peer-focus:-top-2.5 peer-focus:text-sm peer-focus:text-blue-600">
                      Last Name
                    </label>
                  </div>
                </div>

                <div className="relative">
                  <input
                    id="passport"
                    type="text"
                    required
                    className="peer w-full px-4 py-3 border-2 border-gray-200 rounded-lg focus:ring-2 focus:ring-blue-500 focus:border-transparent transition-all placeholder-transparent"
                    placeholder="Passport Number"
                    value={formData.passport}
                    onChange={e => setFormData({...formData, passport: e.target.value})}
                  />
                  <label htmlFor="passport" className="pointer-events-none absolute left-4 -top-2.5 bg-white px-2 text-sm font-medium text-gray-600 transition-all peer-placeholder-shown:top-3 peer-placeholder-shown:text-base peer-placeholder-shown:text-gray-400 peer-focus:-top-2.5 peer-focus:text-sm peer-focus:text-blue-600">
                    Passport Number
                  </label>
                </div>

                <div>
                  <div className="relative">
                    <input
                      type="number"
                      step="0.1"
                      required
                      className="peer w-full px-4 py-3 border-2 border-gray-200 rounded-lg focus:ring-2 focus:ring-blue-500 focus:border-transparent transition-all placeholder-transparent"
                      placeholder="Baggage Weight"
                      value={formData.baggageWeight}
                      onChange={e => setFormData({...formData, baggageWeight: e.target.value})}
                    />
                    <label className="absolute left-4 -top-2.5 bg-white px-2 text-sm font-medium text-gray-600 transition-all peer-placeholder-shown:top-3 peer-placeholder-shown:text-base peer-placeholder-shown:text-gray-400 peer-focus:-top-2.5 peer-focus:text-sm peer-focus:text-blue-600">
                      Baggage Weight (kg)
                    </label>
                  </div>
                  <p className="mt-2 text-sm text-gray-500 flex items-center">
                    <svg className="w-4 h-4 mr-1" fill="currentColor" viewBox="0 0 20 20">
                      <path fillRule="evenodd" d="M18 10a8 8 0 11-16 0 8 8 0 0116 0zm-7-4a1 1 0 11-2 0 1 1 0 012 0zM9 9a1 1 0 000 2v3a1 1 0 001 1h1a1 1 0 100-2v-3a1 1 0 00-1-1H9z" clipRule="evenodd" />
                    </svg>
                    Maximum 25kg checked free
                  </p>
                </div>

                <button
                  disabled={loading}
                  type="submit"
                  className="w-full bg-gradient-to-r from-blue-600 to-blue-700 hover:from-blue-700 hover:to-blue-800 text-white font-bold py-4 px-6 rounded-lg transition-all duration-200 transform hover:scale-[1.02] disabled:opacity-50 disabled:cursor-not-allowed shadow-lg"
                >
                  {loading ? (
                    <span className="flex items-center justify-center">
                      <svg className="animate-spin h-5 w-5 mr-3" fill="none" viewBox="0 0 24 24">
                        <circle className="opacity-25" cx="12" cy="12" r="10" stroke="currentColor" strokeWidth="4"></circle>
                        <path className="opacity-75" fill="currentColor" d="M4 12a8 8 0 018-8V0C5.373 0 0 5.373 0 12h4zm2 5.291A7.962 7.962 0 014 12H0c0 3.042 1.135 5.824 3 7.938l3-2.647z"></path>
                      </svg>
                      Processing...
                    </span>
                  ) : (
                    'Confirm Booking'
                  )}
                </button>
              </form>
            </div>
          </div>
        )}

        {/* Step 4: Booking Confirmation */}
        {step === 4 && bookingResult && (
          <div className="max-w-2xl mx-auto">
            <div className="bg-white rounded-2xl shadow-xl overflow-hidden">
              {/* Success Banner */}
              <div className="bg-gradient-to-r from-emerald-500 to-green-600 px-8 py-8 text-center text-white">
                <div className="w-16 h-16 bg-white/20 rounded-full flex items-center justify-center mx-auto mb-4">
                  <svg className="w-10 h-10" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                    <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2.5} d="M5 13l4 4L19 7" />
                  </svg>
                </div>
                <h2 className="text-3xl font-bold mb-1">Booking Confirmed!</h2>
                <p className="text-emerald-100">Your flight has been successfully booked</p>
              </div>

              {/* Booking Details */}
              <div className="p-8 space-y-6">
                {/* PNR Card */}
                <div className="bg-blue-50 border-2 border-blue-200 rounded-xl p-5">
                  <div className="flex items-center justify-between">
                    <div>
                      <p className="text-sm font-medium text-blue-600 mb-1">PNR Number</p>
                      <p className="text-3xl font-bold text-blue-900 tracking-widest">{bookingResult.pnr}</p>
                    </div>
                    <button
                      onClick={() => copyToClipboard(bookingResult.pnr, 'pnr')}
                      className="px-4 py-2 bg-blue-600 hover:bg-blue-700 text-white rounded-lg transition-colors flex items-center space-x-2 text-sm font-medium"
                    >
                      {copied === 'pnr' ? (
                        <><svg className="w-4 h-4" fill="none" stroke="currentColor" viewBox="0 0 24 24"><path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M5 13l4 4L19 7" /></svg><span>Copied!</span></>
                      ) : (
                        <><svg className="w-4 h-4" fill="none" stroke="currentColor" viewBox="0 0 24 24"><path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M8 16H6a2 2 0 01-2-2V6a2 2 0 012-2h8a2 2 0 012 2v2m-6 12h8a2 2 0 002-2v-8a2 2 0 00-2-2h-8a2 2 0 00-2 2v8a2 2 0 002 2z" /></svg><span>Copy</span></>
                      )}
                    </button>
                  </div>
                </div>

                {/* Booking Reference Card */}
                <div className="bg-gray-50 border-2 border-gray-200 rounded-xl p-5">
                  <div className="flex items-center justify-between">
                    <div>
                      <p className="text-sm font-medium text-gray-600 mb-1">Booking Reference</p>
                      <p className="text-xl font-bold text-gray-900 tracking-wide">{bookingResult.bookingReference}</p>
                    </div>
                    <button
                      onClick={() => copyToClipboard(bookingResult.bookingReference, 'ref')}
                      className="px-4 py-2 bg-gray-600 hover:bg-gray-700 text-white rounded-lg transition-colors flex items-center space-x-2 text-sm font-medium"
                    >
                      {copied === 'ref' ? (
                        <><svg className="w-4 h-4" fill="none" stroke="currentColor" viewBox="0 0 24 24"><path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M5 13l4 4L19 7" /></svg><span>Copied!</span></>
                      ) : (
                        <><svg className="w-4 h-4" fill="none" stroke="currentColor" viewBox="0 0 24 24"><path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M8 16H6a2 2 0 01-2-2V6a2 2 0 012-2h8a2 2 0 012 2v2m-6 12h8a2 2 0 002-2v-8a2 2 0 00-2-2h-8a2 2 0 00-2 2v8a2 2 0 002 2z" /></svg><span>Copy</span></>
                      )}
                    </button>
                  </div>
                </div>

                {/* Flight Summary */}
                <div className="border-t border-gray-200 pt-5">
                  <h3 className="text-sm font-semibold text-gray-500 uppercase tracking-wider mb-3">Flight Details</h3>
                  <div className="grid grid-cols-2 gap-4 text-sm">
                    <div>
                      <p className="text-gray-500">Flight</p>
                      <p className="font-semibold text-gray-900">{bookingResult.flight?.code || bookingResult.flight?.flight_code}</p>
                    </div>
                    <div>
                      <p className="text-gray-500">Seat</p>
                      <p className="font-semibold text-gray-900">{bookingResult.seatNo}</p>
                    </div>
                    <div>
                      <p className="text-gray-500">Passenger</p>
                      <p className="font-semibold text-gray-900">{bookingResult.passenger}</p>
                    </div>
                    <div>
                      <p className="text-gray-500">Route</p>
                      <p className="font-semibold text-gray-900">{bookingResult.flight?.source} → {bookingResult.flight?.destination}</p>
                    </div>
                  </div>
                </div>

                {/* Action Buttons */}
                <div className="space-y-3 pt-2">
                  <button
                    onClick={() => navigate(`/web-check-in?pnr=${bookingResult.pnr}&lastName=${encodeURIComponent(bookingResult.lastName)}`)}
                    className="w-full bg-gradient-to-r from-emerald-600 to-green-600 hover:from-emerald-700 hover:to-green-700 text-white font-bold py-4 px-6 rounded-xl transition-all duration-200 transform hover:scale-[1.02] shadow-lg flex items-center justify-center space-x-3"
                  >
                    <svg className="w-6 h-6" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                      <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M9 5H7a2 2 0 00-2 2v12a2 2 0 002 2h10a2 2 0 002-2V7a2 2 0 00-2-2h-2M9 5a2 2 0 002 2h2a2 2 0 002-2M9 5a2 2 0 012-2h2a2 2 0 012 2m-6 9l2 2 4-4" />
                    </svg>
                    <span>Proceed to Web Check-In</span>
                  </button>

                  <button
                    onClick={() => {
                      setStep(1);
                      setBookingResult(null);
                      setSelectedSeat(null);
                      setSelectedFlight(null);
                      setFormData({ firstName: '', lastName: '', passport: '', baggageWeight: 0 });
                      getFlights().then(data => {
                        setFlights(data);
                        filterFlightsByDate(data, selectedDate);
                      }).catch(console.error);
                    }}
                    className="w-full bg-white border-2 border-gray-300 hover:border-blue-500 text-gray-700 hover:text-blue-700 font-semibold py-3 px-6 rounded-xl transition-all duration-200 flex items-center justify-center space-x-2"
                  >
                    <svg className="w-5 h-5" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                      <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M12 6v6m0 0v6m0-6h6m-6 0H6" />
                    </svg>
                    <span>Book Another Flight</span>
                  </button>
                </div>
              </div>
            </div>
          </div>
        )}

      </main>
    </div>
  );
}

export default App;
