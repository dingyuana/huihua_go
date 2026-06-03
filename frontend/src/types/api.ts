/** 统一 API 响应格式 */
export interface ApiResponse<T = unknown> {
  code: number
  message: string
  data: T
}

/** 分页请求参数 */
export interface PageQuery {
  page: number
  pageSize: number
  sort?: string
}

/** 分页响应 */
export interface PageResult<T> {
  list: T[]
  total: number
  page: number
  pageSize: number
}
