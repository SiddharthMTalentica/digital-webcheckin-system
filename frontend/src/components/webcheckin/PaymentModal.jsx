import { useState } from 'react';
import { useNavigate } from 'react-router-dom';
import { AlertCircle, CreditCard } from 'lucide-react';

export default function PaymentModal({ checkInId, feeDetails, booking, seatNo, onPaymentSuccess }) {
  const [loading, setLoading] = useState(false);
  const [error, setError] = useState(null);

  const handlePayment = async () => {
    setLoading(true);
    setError(null);

    try {
      // Process payment
      const paymentResponse = await fetch(
        `http://localhost:8081/api/webcheckin/${booking.pnr}/baggage-payment`,
        {
          method: 'POST',
          headers: { 'Content-Type': 'application/json' },
          body: JSON.stringify({
            checkInId,
            feeAmount: feeDetails.feeAmount,
            paymentMethod: 'SIMULATED',
          }),
        }
      );

      const paymentData = await paymentResponse.json();

      if (!paymentResponse.ok) {
        throw new Error(paymentData.error || 'Payment failed');
      }

      // Complete check-in
      const checkinResponse = await fetch(
        `http://localhost:8081/api/webcheckin/${booking.pnr}/complete`,
        {
          method: 'POST',
          headers: { 'Content-Type': 'application/json' },
          body: JSON.stringify({
            seatNo,
            baggageWeight: feeDetails.currentWeight,
          }),
        }
      );

      const checkinData = await checkinResponse.json();

      if (!checkinResponse.ok) {
        throw new Error(checkinData.error || 'Check-in failed');
      }

      onPaymentSuccess(checkinData.boardingPass);
    } catch (err) {
      setError(err.message);
    } finally {
      setLoading(false);
    }
  };

  return (
    <div className="fixed inset-0 bg-black bg-opacity-50 flex items-center justify-center p-4 z-50">
      <div className="bg-white rounded-2xl shadow-2xl max-w-md w-full p-8">
        {/* Header */}
        <div className="text-center mb-6">
          <div className="bg-gradient-to-br from-orange-500 to-red-600 p-4 rounded-2xl inline-flex mb-4">
            <CreditCard className="w-8 h-8 text-white" />
          </div>
          <h2 className="text-2xl font-bold text-gray-900 mb-2">
            Baggage Fee Payment
          </h2>
          <p className="text-gray-600">
            Your baggage exceeds the free limit
          </p>
        </div>

        {/* Fee Breakdown */}
        <div className="bg-orange-50 rounded-lg p-4 mb-6">
          <div className="space-y-2 text-sm">
            <div className="flex justify-between">
              <span className="text-gray-600">Bag weight:</span>
              <span className="font-semibold text-gray-900">{feeDetails.currentWeight} kg</span>
            </div>
            <div className="flex justify-between">
              <span className="text-gray-600">Free allowance:</span>
              <span className="font-semibold text-gray-900">{feeDetails.maxFree} kg</span>
            </div>
            <div className="flex justify-between border-t border-orange-200 pt-2">
              <span className="text-gray-600">Excess weight:</span>
              <span className="font-semibold text-orange-600">{feeDetails.excessWeight} kg</span>
            </div>
            <div className="flex justify-between items-center border-t border-orange-200 pt-2">
              <span className="font-semibold text-gray-900">Amount to pay:</span>
              <span className="text-2xl font-bold text-orange-600">
                ${feeDetails.feeAmount.toFixed(2)}
              </span>
            </div>
          </div>
        </div>

        {/* Payment Info */}
        <div className="bg-blue-50 border-2 border-blue-200 rounded-lg p-4 mb-6">
          <p className="text-sm text-blue-800">
            ℹ️ <strong>Demo Mode:</strong> This is a simulated payment. No actual charge will be made.
          </p>
        </div>

        {/* Error */}
        {error && (
          <div className="bg-red-50 border-2 border-red-200 rounded-lg p-4 mb-6 flex items-start gap-3">
            <AlertCircle className="w-5 h-5 text-red-600 flex-shrink-0" />
            <div className="text-red-800 text-sm">{error}</div>
          </div>
        )}

        {/* Actions */}
        <button
          onClick={handlePayment}
          disabled={loading}
          className="w-full bg-gradient-to-r from-orange-500 to-red-600 text-white font-semibold py-4 rounded-lg hover:from-orange-600 hover:to-red-700 transition-all duration-200 shadow-lg hover:shadow-xl disabled:opacity-50 disabled:cursor-not-allowed"
        >
          {loading ? (
            <span className="flex items-center justify-center">
              <svg className="animate-spin -ml-1 mr-3 h-5 w-5 text-white" xmlns="http://www.w3.org/2000/svg" fill="none" viewBox="0 0 24 24">
                <circle className="opacity-25" cx="12" cy="12" r="10" stroke="currentColor" strokeWidth="4"></circle>
                <path className="opacity-75" fill="currentColor" d="M4 12a8 8 0 018-8V0C5.373 0 0 5.373 0 12h4zm2 5.291A7.962 7.962 0 014 12H0c0 3.042 1.135 5.824 3 7.938l3-2.647z"></path>
              </svg>
              Processing...
            </span>
          ) : (
            `Pay $${feeDetails.feeAmount.toFixed(2)} & Complete Check-In`
          )}
        </button>
      </div>
    </div>
  );
}
