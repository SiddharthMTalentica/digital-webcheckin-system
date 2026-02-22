import requests
import json
import time

BACKEND_URL = "http://localhost:8080/api/v1"
WEBCHECKIN_URL = "http://localhost:8081/api/webcheckin"

def verify():
    print("1. Getting Flights from Backend...")
    try:
        resp = requests.get(f"{BACKEND_URL}/flights")
        resp.raise_for_status()
        flights = resp.json()
        if not flights:
            print("No flights found.")
            return
        flight = flights[0]
        flight_id = flight['id']
        print(f"   Found flight: {flight['code']} ({flight_id})")
    except Exception as e:
        print(f"FAIL: Getting flights failed: {e}")
        return

    print("\n2. Getting Available Seats...")
    try:
        resp = requests.get(f"{BACKEND_URL}/flights/{flight_id}/seats")
        resp.raise_for_status()
        data = resp.json()
        seats = data['seats']
        available_seat = next((s for s in seats if s['status'] == 'AVAILABLE'), None)
        if not available_seat:
            print("No available seats.")
            return
        seat_no = available_seat['seat_no']
        print(f"   Found available seat: {seat_no}")
    except Exception as e:
        print(f"FAIL: Getting seats failed: {e}")
        return

    print("\n3. Holding Seat...")
    user_id = "test-user-123"
    try:
        resp = requests.post(f"{BACKEND_URL}/flights/{flight_id}/seats/{seat_no}/hold", json={"user_id": user_id})
        resp.raise_for_status()
        print("   Seat held successfully.")
    except Exception as e:
        print(f"FAIL: Holding seat failed: {e}")
        # Proceeding anyway as check-in might work if hold is technically not required by some logic or race condition, but usually it is.
        # actually confirm checkin requires hold.
    
    print("\n4. Confirming Booking (Backend)...")
    pnr = ""
    last_name = "Doe"
    try:
        payload = {
            "flight_id": flight_id,
            "seat_no": seat_no,
            "user_id": user_id,
            "first_name": "John",
            "last_name": last_name,
            "passport": "P12345678",
            "baggage_weight": 15.0
        }
        resp = requests.post(f"{BACKEND_URL}/checkin/confirm", json=payload) # Note: path is /checkin/confirm based on handler.Routes
        resp.raise_for_status()
        booking = resp.json()
        print(f"   Booking created: {json.dumps(booking, indent=2)}")
        
        if 'pnr' in booking:
            pnr = booking['pnr']
            print(f"   Captured PNR: {pnr}")
        else:
            print("FAIL: PNR not found in booking response.")
            return

    except Exception as e:
        print(f"FAIL: Confirming booking failed: {e}")
        if resp:
            print(f"Response: {resp.text}")
        return

    print("\n5. Validating with Web Check-in Service...")
    try:
        # Looking up verify endpoint. backend_webcheckin/cmd/main.go says:
        # api.Post("/lookup", checkInHandler.LookupBooking)
        
        lookup_payload = {
            "pnr": pnr,
            "lastName": last_name
        }
        resp = requests.post(f"{WEBCHECKIN_URL}/lookup", json=lookup_payload)
        
        if resp.status_code == 200:
            print("SUCCESS: Web Check-in Lookup Successful!")
            print(json.dumps(resp.json(), indent=2))
        else:
            print(f"FAIL: Web Check-in Lookup Failed: {resp.status_code}")
            print(resp.text)

    except Exception as e:
        print(f"FAIL: Web Check-in request failed: {e}")

if __name__ == "__main__":
    verify()
