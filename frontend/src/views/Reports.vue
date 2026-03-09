<template>
  <div class="reports-page">
    <h2 class="page-title">销售报表</h2>

    <div class="filter-bar card">
      <div class="form-group">
        <label class="form-label">开始日期</label>
        <input v-model="filter.startDate" type="date" class="form-input" @change="fetchData" />
      </div>
      <div class="form-group">
        <label class="form-label">结束日期</label>
        <input v-model="filter.endDate" type="date" class="form-input" @change="fetchData" />
      </div>
      <div class="quick-buttons">
        <button class="btn btn-small btn-secondary" @click="setDateRange(7)">近7天</button>
        <button class="btn btn-small btn-secondary" @click="setDateRange(30)">近30天</button>
        <button class="btn btn-small btn-secondary" @click="setDateRange(90)">近90天</button>
      </div>
    </div>

    <div class="charts-grid">
      <div class="card">
        <h3 class="card-title">销售趋势</h3>
        <div ref="salesChartRef" class="chart-container"></div>
      </div>

      <div class="card">
        <h3 class="card-title">订单数量趋势</h3>
        <div ref="ordersChartRef" class="chart-container"></div>
      </div>
    </div>

    <div class="card">
      <h3 class="card-title">商品销量排行 TOP 10</h3>
      <table class="table">
        <thead>
          <tr>
            <th>排名</th>
            <th>商品名称</th>
            <th>销量</th>
            <th>销售额</th>
          </tr>
        </thead>
        <tbody>
          <tr v-for="(item, index) in productRank" :key="item.product_id">
            <td>
              <span class="rank-badge" :class="{ top: index < 3 }">{{ index + 1 }}</span>
            </td>
            <td>{{ item.product_name }}</td>
            <td>{{ item.sales }} 件</td>
            <td class="price">¥{{ item.amount.toFixed(2) }}</td>
          </tr>
        </tbody>
      </table>
      <div v-if="productRank.length === 0" class="empty-state">
        暂无数据
      </div>
    </div>

    <div class="stats-summary card">
      <h3 class="card-title">统计汇总</h3>
      <div class="summary-grid">
        <div class="summary-item">
          <div class="summary-value">{{ summary.totalOrders }}</div>
          <div class="summary-label">总订单数</div>
        </div>
        <div class="summary-item">
          <div class="summary-value">¥{{ summary.totalAmount.toFixed(2) }}</div>
          <div class="summary-label">总销售额</div>
        </div>
        <div class="summary-item">
          <div class="summary-value">{{ summary.totalProducts }}</div>
          <div class="summary-label">售出商品数</div>
        </div>
        <div class="summary-item">
          <div class="summary-value">¥{{ summary.avgOrder.toFixed(2) }}</div>
          <div class="summary-label">平均客单价</div>
        </div>
      </div>
    </div>
  </div>
</template>

<script setup>
import { ref, reactive, computed, onMounted, onUnmounted } from 'vue'
import * as echarts from 'echarts'
import { reportAPI } from '../api'

const salesChartRef = ref(null)
const ordersChartRef = ref(null)
let salesChart = null
let ordersChart = null

const salesData = ref([])
const productRank = ref([])

const filter = reactive({
  startDate: '',
  endDate: ''
})

const summary = computed(() => {
  const totalOrders = salesData.value.reduce((sum, d) => sum + d.orders, 0)
  const totalAmount = salesData.value.reduce((sum, d) => sum + d.amount, 0)
  const totalProducts = productRank.value.reduce((sum, p) => sum + p.sales, 0)
  return {
    totalOrders,
    totalAmount,
    totalProducts,
    avgOrder: totalOrders > 0 ? totalAmount / totalOrders : 0
  }
})

const setDateRange = (days) => {
  const endDate = new Date()
  const startDate = new Date(Date.now() - days * 24 * 60 * 60 * 1000)

  filter.endDate = endDate.toISOString().split('T')[0]
  filter.startDate = startDate.toISOString().split('T')[0]

  fetchData()
}

const fetchData = async () => {
  if (!filter.startDate || !filter.endDate) return

  try {
    const [sales, rank] = await Promise.all([
      reportAPI.getSales({ start_date: filter.startDate, end_date: filter.endDate }),
      reportAPI.getProductRank({ start_date: filter.startDate, end_date: filter.endDate, limit: 10 })
    ])

    salesData.value = sales || []
    productRank.value = rank || []

    renderCharts()
  } catch (error) {
    console.error('获取报表数据失败:', error)
  }
}

const renderCharts = () => {
  renderSalesChart()
  renderOrdersChart()
}

const renderSalesChart = () => {
  if (!salesChartRef.value) return

  if (!salesChart) {
    salesChart = echarts.init(salesChartRef.value)
  }

  const option = {
    tooltip: {
      trigger: 'axis',
      formatter: (params) => {
        const data = params[0]
        return `${data.name}<br/>销售额: ¥${data.value.toFixed(2)}`
      }
    },
    grid: {
      left: '3%',
      right: '4%',
      bottom: '3%',
      containLabel: true
    },
    xAxis: {
      type: 'category',
      boundaryGap: false,
      data: salesData.value.map(d => d.date)
    },
    yAxis: {
      type: 'value',
      axisLabel: {
        formatter: '¥{value}'
      }
    },
    series: [{
      name: '销售额',
      type: 'line',
      smooth: true,
      areaStyle: {
        color: new echarts.graphic.LinearGradient(0, 0, 0, 1, [
          { offset: 0, color: 'rgba(212, 165, 116, 0.5)' },
          { offset: 1, color: 'rgba(212, 165, 116, 0.05)' }
        ])
      },
      lineStyle: { color: '#d4a574', width: 2 },
      itemStyle: { color: '#d4a574' },
      data: salesData.value.map(d => d.amount)
    }]
  }

  salesChart.setOption(option)
}

const renderOrdersChart = () => {
  if (!ordersChartRef.value) return

  if (!ordersChart) {
    ordersChart = echarts.init(ordersChartRef.value)
  }

  const option = {
    tooltip: {
      trigger: 'axis'
    },
    grid: {
      left: '3%',
      right: '4%',
      bottom: '3%',
      containLabel: true
    },
    xAxis: {
      type: 'category',
      data: salesData.value.map(d => d.date)
    },
    yAxis: {
      type: 'value'
    },
    series: [{
      name: '订单数',
      type: 'bar',
      barWidth: '60%',
      itemStyle: {
        color: new echarts.graphic.LinearGradient(0, 0, 0, 1, [
          { offset: 0, color: '#8b7355' },
          { offset: 1, color: '#d4a574' }
        ])
      },
      data: salesData.value.map(d => d.orders)
    }]
  }

  ordersChart.setOption(option)
}

onMounted(() => {
  setDateRange(30)

  window.addEventListener('resize', () => {
    salesChart?.resize()
    ordersChart?.resize()
  })
})

onUnmounted(() => {
  salesChart?.dispose()
  ordersChart?.dispose()
})
</script>

<style scoped>
.page-title {
  font-size: 24px;
  margin-bottom: 24px;
}

.filter-bar {
  display: flex;
  align-items: flex-end;
  gap: 16px;
  margin-bottom: 24px;
}

.filter-bar .form-group {
  margin-bottom: 0;
}

.quick-buttons {
  display: flex;
  gap: 8px;
  margin-left: auto;
}

.charts-grid {
  display: grid;
  grid-template-columns: 1fr 1fr;
  gap: 20px;
  margin-bottom: 24px;
}

.chart-container {
  height: 300px;
}

.rank-badge {
  display: inline-flex;
  align-items: center;
  justify-content: center;
  width: 24px;
  height: 24px;
  border-radius: 50%;
  background: #f0f0f0;
  font-size: 12px;
}

.rank-badge.top {
  background: var(--primary-color);
  color: #fff;
}

.summary-grid {
  display: grid;
  grid-template-columns: repeat(4, 1fr);
  gap: 20px;
}

.summary-item {
  text-align: center;
  padding: 20px;
  background: #fafafa;
  border-radius: 8px;
}

.summary-value {
  font-size: 24px;
  font-weight: 600;
  color: var(--primary-color);
  margin-bottom: 8px;
}

.summary-label {
  color: var(--text-secondary);
}

@media (max-width: 1200px) {
  .charts-grid {
    grid-template-columns: 1fr;
  }

  .summary-grid {
    grid-template-columns: repeat(2, 1fr);
  }
}
</style>
