<template>
  <div>
    <n-h1>凭证管理</n-h1>
    <n-space vertical>
      <n-button type="primary" @click="$router.push('/vouchers/new')">新增凭证</n-button>
      <n-data-table
        :columns="columns"
        :data="data"
        :pagination="pagination"
        :loading="loading"
      />
    </n-space>
  </div>
</template>

<script setup lang="ts">
import { ref, onMounted } from 'vue'
import { NButton, NTag, NDataTable, NText } from 'naive-ui'
import { getVoucherList } from '@/api/voucher'
import { VOUCHER_STATUS_MAP } from '@/types'
import type { Voucher } from '@/types'

const loading = ref(false)
const data = ref<Voucher[]>([])
const pagination = { pageSize: 20 }

const columns = [
  { title: '凭证号', key: 'number', width: 150 },
  { title: '日期', key: 'date', width: 120 },
  { title: '状态', key: 'status', width: 100,
    render(row: any) {
      const s = VOUCHER_STATUS_MAP[row.status] || { label: '未知', color: 'default' }
      return h(NTag, { type: s.color as any }, () => s.label)
    }
  },
  { title: '分录数', key: 'lineCount', width: 80 },
  { title: '金额', key: 'amount', width: 150,
    render(row: any) { return `¥ ${(row.amount || 0).toLocaleString()}` }
  },
  { title: '操作', key: 'actions', width: 150,
    render(row: any) {
      return h(NButton, { size: 'small', onClick: () => $router.push(`/vouchers/${row.id}`) }, () => '查看')
    }
  }
]

import { h } from 'vue'

onMounted(async () => {
  loading.value = true
  try {
    data.value = await getVoucherList()
  } finally {
    loading.value = false
  }
})
</script>