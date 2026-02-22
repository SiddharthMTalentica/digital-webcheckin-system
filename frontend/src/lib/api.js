import axios from 'axios';

const API_BASE_URL = 'http://localhost:8080/api/v1';

const api = axios.create({
  baseURL: API_BASE_URL,
  headers: {
    'Content-Type': 'application/json',
  },
});

export const getFlights = async () => {
    const response = await api.get('/flights');
    return response.data;
};

export const getSeatMap = async (flightId) => {
    const response = await api.get(`/flights/${flightId}/seats`);
    return response.data;
};

export const holdSeat = async (flightId, seatNo, userId) => {
    const response = await api.post(`/flights/${flightId}/seats/${seatNo}/hold`, { user_id: userId });
    return response.data;
};

export const confirmCheckIn = async (data) => {
    const response = await api.post('/checkin/confirm', data);
    return response.data;
};

export default api;
