import { useState } from 'react';
import { useNavigate } from 'react-router-dom';
import { AlertCircle, Plane } from 'lucide-react';

export default function PNRLookup({ onLookupSuccess, initialPnr = '', initialLastName = '' }) {
  const navigate = useNavigate();
  const [pnr, setPnr] = useState(initialPnr);
  const [lastName, setLastName] = useState(initialLastName);
  const [error, setError] = useState(null);
  const [loading, setLoading] = useState(false);

  const handleSubmit = async (e) => {
    e.preventDefault();
    setError(null);

    if (pnr.length !== 6) {
      setError('PNR must be 6 characters');
      return;
    }

    if (!lastName.trim()) {
      setError('Last name is required');
      return;
    }

    setLoading(true);

    try {
      const response = await fetch('http://localhost:8081/api/webcheckin/lookup', {
        method: 'POST',
        headers: { 'Content-Type': 'application/json' },
        body: JSON.stringify({ pnr: pnr.toUpperCase(), lastName }),
      });

      const data = await response.json();

      if (!response.ok) {
        throw new Error(data.error || 'Booking not found');
      }

      // Check if already checked in
      if (data.checkinStatus === 'COMPLETED') {
        setError('You have already checked in for this flight');
        return;
      }

      onLookupSuccess(data.booking);
    } catch (err) {
      setError(err.message);
    } finally {
      setLoading(false);
    }
  };

  return (
    <div className="min-h-screen bg-gradient-to-br from-green-50 via-emerald-50 to-teal-50 flex items-center justify-center p-4">
      {/* Home Button */}
      <button
        onClick={() => navigate('/')}
        className="fixed top-4 left-4 px-4 py-2 bg-white/90 hover:bg-white shadow-lg rounded-lg transition-all duration-200 flex items-center space-x-2 text-gray-700 hover:text-blue-700 z-10"
      >
        <svg className="w-5 h-5" fill="none" stroke="currentColor" viewBox="0 0 24 24">
          <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M3 12l2-2m0 0l7-7 7 7M5 10v10a1 1 0 001 1h3m10-11l2 2m-2-2v10a1 1 0 01-1 1h-3m-6 0a1 1 0 001-1v-4a1 1 0 011-1h2a1 1 0 011 1v4a1 1 0 001 1m-6 0h6" />
        </svg>
        <span className="font-medium">Home</span>
      </button>

      <div className="max-w-md w-full">
        {/* Header */}
        <div className="text-center mb-8">
          <div className="flex items-center justify-center mb-4">
            <div className="bg-gradient-to-br from-green-500 to-emerald-600 p-4 rounded-2xl">
              <Plane className="w-10 h-10 text-white" />
            </div>
          </div>
          <h1 className="text-3xl font-bold text-gray-900 mb-2">
            Web Check-In
          </h1>
          <p className="text-gray-600">
            Enter your booking details to begin
          </p>
        </div>

        {/* Form Card */}
        <div className="bg-white rounded-2xl shadow-xl p-8">
          <form onSubmit={handleSubmit} className="space-y-6">
            {/* PNR Input */}
            <div>
              <label htmlFor="pnr" className="block text-sm font-semibold text-gray-700 mb-2">
                PNR / Booking Reference
              </label>
              <input
                id="pnr"
                type="text"
                maxLength="6"
                value={pnr}
                onChange={(e) => setPnr(e.target.value.toUpperCase())}
                placeholder="ABC123"
                className="w-full px-4 py-3 rounded-lg border-2 border-gray-200 focus:border-green-500 focus:ring-2 focus:ring-green-200 outline-none transition-all font-mono text-lg tracking-wider uppercase"
                required
              />
              <p className="text-xs text-gray-500 mt-1">6-character code from your booking confirmation</p>
            </div>

            {/* Last Name Input */}
            <div>
              <label htmlFor="lastName" className="block text-sm font-semibold text-gray-700 mb-2">
                Last Name
              </label>
              <input
                id="lastName"
                type="text"
                value={lastName}
                onChange={(e) => setLastName(e.target.value)}
                placeholder="Smith"
                className="w-full px-4 py-3 rounded-lg border-2 border-gray-200 focus:border-green-500 focus:ring-2 focus:ring-green-200 outline-none transition-all"
                required
              />
              <p className="text-xs text-gray-500 mt-1">As shown on your booking</p>
            </div>

            {/* Error Message */}
            {error && (
              <div className="bg-red-50 border-2 border-red-200 rounded-lg p-4 flex items-start gap-3">
                <AlertCircle className="w-5 h-5 text-red-600 flex-shrink-0 mt-0.5" />
                <div className="text-red-800 text-sm">{error}</div>
              </div>
            )}

            {/* Submit Button */}
            <button
              type="submit"
              disabled={loading}
              className="w-full bg-gradient-to-r from-green-500 to-emerald-600 text-white font-semibold py-4 rounded-lg hover:from-green-600 hover:to-emerald-700 transition-all duration-200 shadow-lg hover:shadow-xl disabled:opacity-50 disabled:cursor-not-allowed"
            >
              {loading ? (
                <span className="flex items-center justify-center">
                  <svg className="animate-spin -ml-1 mr-3 h-5 w-5 text-white" xmlns="http://www.w3.org/2000/svg" fill="none" viewBox="0 0 24 24">
                    <circle className="opacity-25" cx="12" cy="12" r="10" stroke="currentColor" strokeWidth="4"></circle>
                    <path className="opacity-75" fill="currentColor" d="M4 12a8 8 0 018-8V0C5.373 0 0 5.373 0 12h4zm2 5.291A7.962 7.962 0 014 12H0c0 3.042 1.135 5.824 3 7.938l3-2.647z"></path>
                  </svg>
                  Verifying...
                </span>
              ) : (
                'Continue to Check-In'
              )}
            </button>
          </form>

          {/* Sample PNRs */}
          <div className="mt-6 pt-6 border-t border-gray-200">
            <p className="text-xs text-gray-500 text-center mb-2">Test PNRs for demo:</p>
            <div className="flex gap-2 justify-center flex-wrap">
              <button
                type="button"
                onClick={() => { setPnr('ABC123'); setLastName('Doe'); }}
                className="text-xs bg-gray-100 hover:bg-gray-200 px-3 py-1 rounded-full transition-colors"
              >
                ABC123 / Doe
              </button>
              <button
                type="button"
                onClick={() => { setPnr('XYZ789'); setLastName('Smith'); }}
                className="text-xs bg-gray-100 hover:bg-gray-200 px-3 py-1 rounded-full transition-colors"
              >
                XYZ789 / Smith
              </button>
            </div>
          </div>
        </div>

        {/* Help Text */}
        <p className="text-center text-sm text-gray-500 mt-6">
          Don't have a booking? <a href="/flight-booking-system" className="text-green-600 hover:underline font-semibold">Book a flight</a>
        </p>
      </div>
    </div>
  );
}
