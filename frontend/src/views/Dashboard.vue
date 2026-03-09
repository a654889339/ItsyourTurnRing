<template>
  <div class="dashboard">
    <h2 class="page-title">仪表盘</h2>

    <div class="stats-grid">
      <div class="stat-card">
        <div class="stat-icon">📦</div>
        <div class="stat-info">
          <div class="stat-value">{{ stats.today_orders }}</div>
          <div class="stat-label">今日订单</div>
        </div>
      </div>

      <div class="stat-card">
        <div class="stat-icon">💰</div>
        <div class="stat-info">
          <div class="stat-value">¥{{ stats.today_amount?.toFixed(2) || '0.00' }}</div>
          <div class="stat-label">今日收入</div>
        </div>
      </div>

      <div class="stat-card">
        <div class="stat-icon">💍</div>
        <div class="stat-info">
          <div class="stat-value">{{ stats.total_products }}</div>
          <div class="stat-label">商品总数</div>
        </div>
      </div>

      <div class="stat-card">
        <div class="stat-icon">⏳</div>
        <div class="stat-info">
          <div class="stat-value">{{ stats.pending_orders }}</div>
          <div class="stat-label">待处理订单</div>
        </div>
      </div>
    </div>

    <div class="charts-row">
      <div class="card chart-card">
        <h3 class="card-title">销售趋势</h3>
        <div ref="salesChartRef" class="chart-container"></div>
      </div>

      <div class="card chart-card">
        <h3 class="card-title">商品销量排行</h3>
        <div class="rank-list">
          <div v-for="(item, index) in productRank" :key="item.product_id" class="rank-item">
            <span class="rank-num" :class="{ top: index < 3 }">{{ index + 1 }}</span>
            <span class="rank-name">{{ item.product_name }}</span>
            <span class="rank-sales">{{ item.sales }}件</span>
          </div>
          <div v-if="productRank.length === 0" class="empty-state">
            暂无数据
          </div>
        </div>
      </div>
    </div>

    <div class="card" v-if="stats.low_stock_products > 0">
      <h3 class="card-title">库存预警</h3>
      <p class="warning-text">有 {{ stats.low_stock_products }} 个商品库存不足10件</p>
    </div>
  </div>
</template>

<script setup>
import { ref, onMounted, onUnmounted } from 'vue'
import * as echarts from 'echarts'
import { reportAPI } from '../api'

const stats = ref({})
const productRank = ref([])
const salesChartRef = ref(null)
let salesChart = null

const fetchDashboard = async () => {
  try {
    stats.value = await reportAPI.getDashboard()
  } catch (error) {
    console.error('获取仪表盘数据失败:', error)
  }
}

const fetchSalesData = async () => {
  try {
    const endDate = new Date().toISOString().split('T')[0]
    const startDate = new Date(Date.now() - 7 * 24 * 60 * 60 * 1000).toISOString().split('T')[0]

    const data = await reportAPI.getSales({ start_date: startDate, end_date: endDate })
    renderSalesChart(data || [])
  } catch (error) {
    console.error('获取销售数据失败:', error)
  }
}

const fetchProductRank = async () => {
  try {
    const endDate = new Date().toISOString().split('T')[0]
    const startDate = new Date(Date.now() - 30 * 24 * 60 * 60 * 1000).toISOString().split('T')[0]

    productRank.value = await reportAPI.getProductRank({
      start_date: startDate,
      end_date: endDate,
      limit: 10
    }) || []
  } catch (error) {
    console.error('获取商品排行失败:', error)
  }
}

const renderSalesChart = (data) => {
  if (!salesChartRef.value) return

  if (!salesChart) {
    salesChart = echarts.init(salesChartRef.value)
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
      boundaryGap: false,
      data: data.map(d => d.date)
    },
    yAxis: [
      {
        type: 'value',
        name: '订单数'
      },
      {
        type: 'value',
        name: '金额(元)',
        position: 'right'
      }
    ],
    series: [
      {
        name: '订单数',
        type: 'line',
        smooth: true,
        data: data.map(d => d.orders),
        itemStyle: { color: '#d4a574' }
      },
      {
        name: '金额',
        type: 'bar',
        yAxisIndex: 1,
        data: data.map(d => d.amount),
        itemStyle: { color: '#8b7355' }
      }
    ]
  }

  salesChart.setOption(option)
}

onMounted(() => {
  fetchDashboard()
  fetchSalesData()
  fetchProductRank()

  window.addEventListener('resize', () => {
    salesChart?.resize()
  })
})

onUnmounted(() => {
  salesChart?.dispose()
})
</script>

<style scoped>
.page-title {
  font-size: 24px;
  margin-bottom: 24px;
}

.stats-grid {
  display: grid;
  grid-template-columns: repeat(4, 1fr);
  gap: 20px;
  margin-bottom: 24px;
}

.stat-card {
  background: #fff;
  border-radius: 8px;
  padding: 20px;
  display: flex;
  align-items: center;
  box-shadow: 0 1px 3px rgba(0, 0, 0, 0.1);
}

.stat-icon {
  font-size: 40px;
  margin-right: 16px;
}

.stat-value {
  font-size: 24px;
  font-weight: 600;
  color: var(--text-color);
}

.stat-label {
  font-size: 14px;
  color: var(--text-secondary);
}

.charts-row {
  display: grid;
  grid-template-columns: 2fr 1fr;
  gap: 20px;
  margin-bottom: 24px;
}

.chart-card {
  min-height: 350px;
}

.chart-container {
  height: 280px;
}

.rank-list {
  max-height: 280px;
  overflow-y: auto;
}

.rank-item {
  display: flex;
  align-items: center;
  padding: 10px 0;
  border-bottom: 1px solid var(--border-color);
}

.rank-num {
  width: 24px;
  height: 24px;
  border-radius: 50%;
  background: var(--border-color);
  display: flex;
  align-items: center;
  justify-content: center;
  font-size: 12px;
  margin-right: 12px;
}

.rank-num.top {
  background: var(--primary-color);
  color: #fff;
}

.rank-name {
  flex: 1;
  overflow: hidden;
  text-overflow: ellipsis;
  white-space: nowrap;
}

.rank-sales {
  color: var(--text-secondary);
  font-size: 13px;
}

.warning-text {
  color: var(--warning-color);
}

@media (max-width: 1200px) {
  .stats-grid {
    grid-template-columns: repeat(2, 1fr);
  }

  .charts-row {
    grid-template-columns: 1fr;
  }
}
</style>
