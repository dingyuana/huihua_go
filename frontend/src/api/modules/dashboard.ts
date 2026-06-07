import request from '../request'

export const fetchDashboardStats = () => request.get('/dashboard/stats')