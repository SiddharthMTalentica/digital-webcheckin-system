import { Link } from 'react-router-dom';
import { Plane, Luggage } from 'lucide-react';

export default function Homepage() {
  return (
    <div className="min-h-screen bg-gradient-to-br from-blue-50 via-indigo-50 to-purple-50 flex items-center justify-center p-4">
      <div className="max-w-6xl w-full">
        {/* Header */}
        <div className="text-center mb-12">
          <div className="flex items-center justify-center mb-4">
            <Plane className="w-16 h-16 text-indigo-600" />
          </div>
          <h1 className="text-5xl font-bold text-gray-900 mb-3">
            SkyHigh Airlines
          </h1>
          <p className="text-xl text-gray-600">
            Welcome! Choose your service
          </p>
        </div>

        {/* Service Cards */}
        <div className="grid md:grid-cols-2 gap-8">
          {/* Flight Booking */}
          <Link to="/flight-booking-system">
            <div className="group bg-white rounded-2xl shadow-lg hover:shadow-2xl transition-all duration-300 p-8 border-2 border-transparent hover:border-indigo-500 cursor-pointer">
              <div className="flex flex-col items-center text-center">
                <div className="bg-gradient-to-br from-indigo-500 to-purple-600 p-6 rounded-2xl mb-6 group-hover:scale-110 transition-transform duration-300">
                  <Plane className="w-12 h-12 text-white" />
                </div>
                
                <h2 className="text-2xl font-bold text-gray-900 mb-3">
                  Book a Flight
                </h2>
                
                <p className="text-gray-600 mb-6">
                  Search and book flights, select seats, and complete your booking
                </p>
                
                <div className="flex flex-wrap gap-2 justify-center text-sm text-gray-500">
                  <span className="bg-gray-100 px-3 py-1 rounded-full">Search Flights</span>
                  <span className="bg-gray-100 px-3 py-1 rounded-full">Select Seats</span>
                  <span className="bg-gray-100 px-3 py-1 rounded-full">Book Now</span>
                </div>
              </div>
            </div>
          </Link>

          {/* Web Check-In */}
          <Link to="/web-check-in">
            <div className="group bg-white rounded-2xl shadow-lg hover:shadow-2xl transition-all duration-300 p-8 border-2 border-transparent hover:border-green-500 cursor-pointer">
              <div className="flex flex-col items-center text-center">
                <div className="bg-gradient-to-br from-green-500 to-emerald-600 p-6 rounded-2xl mb-6 group-hover:scale-110 transition-transform duration-300">
                  <Luggage className="w-12 h-12 text-white" />
                </div>
                
                <h2 className="text-2xl font-bold text-gray-900 mb-3">
                  Web Check-In
                </h2>
                
                <p className="text-gray-600 mb-6">
                  Already have a booking? Check-in online with your PNR
                </p>
                
                <div className="flex flex-wrap gap-2 justify-center text-sm text-gray-500">
                  <span className="bg-gray-100 px-3 py-1 rounded-full">Enter PNR</span>
                  <span className="bg-gray-100 px-3 py-1 rounded-full">Select Seat</span>
                  <span className="bg-gray-100 px-3 py-1 rounded-full">Get Boarding Pass</span>
                </div>
              </div>
            </div>
          </Link>
        </div>

        {/* Footer Info */}
        <div className="text-center mt-12 text-gray-500 text-sm">
          <p>SkyHigh Core – Digital Check-In System</p>
        </div>
      </div>
    </div>
  );
}
