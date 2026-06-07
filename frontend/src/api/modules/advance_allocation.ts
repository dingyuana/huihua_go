import request from '@/api/request'
import type { ApiResponse } from '@/types/api'

export interface AdvanceAllocation {
  id: string
  advance_id: string
  advance_type: 'receipt' | 'payment'
  target_id: string
  target_type: 'ar' | 'ap'
  allocated_amount: string
  allocation_date: string
  voucher_no?: string
  remark?: string
  created_at: string
}

export interface AllocatePayload {
  advance_id: string
  advance_type: 'receipt' | 'payment'
  target_id: string
  target_type: 'ar' | 'ap'
  allocated_amount: string
  allocation_date: string
  remark?: string
}

export function allocateAdvance(payload: AllocatePayload): Promise<ApiResponse<AdvanceAllocation>> {
  return request.post('/advance-allocations', payload)
}

export function autoMatchAdvance(id: string): Promise<ApiResponse<AdvanceAllocation[]>> {
  return request.post(`/advance-allocations/${id}/auto-match`)
}

export function listAdvanceAllocations(advanceId: string): Promise<ApiResponse<{ list: AdvanceAllocation[]; total: number }>> {
  return request.get('/advance-allocations', { params: { advance_id: advanceId } })
}

export function listAdvanceAllocationsByTarget(targetId: string): Promise<ApiResponse<{ list: AdvanceAllocation[]; total: number }>> {
  return request.get('/advance-allocations/by-target', { params: { target_id: targetId } })
}
