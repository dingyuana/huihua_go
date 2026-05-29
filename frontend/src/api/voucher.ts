import apiClient from './index'
import type { Voucher, VoucherLine } from '@/types'

export async function getVoucherList(params?: { page?: number; page_size?: number; status?: number }) {
  const res = await apiClient.get('/api/v1/vouchers', { params })
  return res.data?.data || res.data || []
}

export async function getVoucherById(id: string) {
  const res = await apiClient.get(`/api/v1/vouchers/${id}`)
  return res.data?.data || res.data
}

export async function createVoucher(payload: {
  posting_date: string
  remark: string
  lines: { account_id: string; debit: number; credit: number; remark?: string }[]
}) {
  const res = await apiClient.post('/api/v1/vouchers', payload)
  return res.data?.data || res.data
}

export async function updateVoucher(id: string, payload: any) {
  const res = await apiClient.put(`/api/v1/vouchers/${id}`, payload)
  return res.data?.data || res.data
}

export async function submitVoucher(id: string) {
  const res = await apiClient.post(`/api/v1/vouchers/${id}/submit`)
  return res.data?.data || res.data
}

export async function approveVoucher(id: string) {
  const res = await apiClient.post(`/api/v1/vouchers/${id}/approve`)
  return res.data?.data || res.data
}

export async function rejectVoucher(id: string, reason?: string) {
  const res = await apiClient.post(`/api/v1/vouchers/${id}/reject`, { reason })
  return res.data?.data || res.data
}

export async function reverseVoucher(id: string) {
  const res = await apiClient.post(`/api/v1/vouchers/${id}/reverse`)
  return res.data?.data || res.data
}

export async function deleteVoucher(id: string) {
  const res = await apiClient.delete(`/api/v1/vouchers/${id}`)
  return res.data
}

export async function getAccounts() {
  const res = await apiClient.get('/api/v1/accounts/tree')
  return res.data?.data || res.data || []
}

export async function getNextVoucherNo() {
  const res = await apiClient.post('/api/v1/voucher-templates/numbering-rule/next')
  return res.data?.data || res.data
}