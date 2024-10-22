<template>
  <div class="nft-collections">
    <h2>NFT 系列</h2>
    <el-row v-if="!loading && !error" :gutter="20">
      <el-col :span="6" v-for="(collection, index) in sortedCollections" :key="index">
        <el-card :body-style="{ padding: '0px' }" shadow="hover" class="collection-card">
          <router-link :to="`/nft/${collection.address}`" class="collection-link">
            <img :src="collection.imageUrl" class="image" :alt="collection.name">
            <div class="collection-info">
              <span class="collection-name">{{ collection.name }}</span>
              <div class="bottom">
                <span v-if="collection.floorPrice" class="floor-price">地板价: {{ formatPrice(collection.floorPrice) }} REX</span>
                <span v-else class="floor-price">暂无出售</span>
              </div>
            </div>
          </router-link>
        </el-card>
      </el-col>
    </el-row>
    <el-row v-if="loading" :gutter="20">
      <el-col :span="6" v-for="i in 4" :key="i">
        <el-card :body-style="{ padding: '0px' }" shadow="hover" class="collection-card">
          <el-skeleton :loading="loading" animated>
            <template #template>
              <el-skeleton-item variant="image" style="width: 100%; height: 200px;" />
              <div style="padding: 14px;">
                <el-skeleton-item variant="h3" style="width: 50%;" />
                <div style="display: flex; align-items: center; justify-content: space-between; margin-top: 13px;">
                  <el-skeleton-item variant="text" style="width: 60%;" />
                </div>
              </div>
            </template>
          </el-skeleton>
        </el-card>
      </el-col>
    </el-row>
    <div v-if="error" class="error">
      {{ error }}
    </div>
  </div>
</template>

<script>
import { ref, onMounted, computed } from 'vue';
import { ElSkeleton, ElSkeletonItem } from 'element-plus';
import { ethers } from 'ethers';
import { getIPFSUrl } from '../utils/nftUtils';
import axios from 'axios';

const API_BASE_URL = 'http://121.196.204.174:8081/api';

export default {
  components: {
    ElSkeleton,
    ElSkeletonItem,
  },
  setup() {
    const collections = ref([]);
    const loading = ref(true);
    const error = ref(null);

    const fetchCollections = async () => {
      try {
        loading.value = true;
        error.value = null;

        // 获取所有 NFT 集合
        const allCollectionsResponse = await axios.get(`${API_BASE_URL}/nft`);
        const allCollections = allCollectionsResponse.data;

        // 获取所有订单
        const ordersResponse = await axios.get(`${API_BASE_URL}/orders`);
        const orders = ordersResponse.data;

        // 处理集合信息
        const processedCollections = allCollections.map(collection => {
          const activeOrders = orders.filter(order => 
            order.NFTContractAddress.toLowerCase() === collection.ContractAddress.toLowerCase() && 
            order.Status === 0  // 0 表示出售中
          );

          let floorPrice = null;
          if (activeOrders.length > 0) {
            floorPrice = activeOrders.reduce((min, order) => {
              // 确保 Price 是一个有效的数字字符串
              const price = order.Price.toString();
              return ethers.BigNumber.from(price).lt(ethers.BigNumber.from(min)) ? price : min;
            }, activeOrders[0].Price.toString());
          }

          return {
            address: collection.ContractAddress,
            name: collection.Name,
            imageUrl: getIPFSUrl(collection.TokenIconURI),
            floorPrice: floorPrice,
            hasActiveOrders: activeOrders.length > 0
          };
        });

        collections.value = processedCollections;
      } catch (err) {
        console.error('获取 NFT 系列失败:', err);
        error.value = '加载 NFT 系列失败，请稍后重试';
      } finally {
        loading.value = false;
      }
    };

    const sortedCollections = computed(() => {
      return [...collections.value].sort((a, b) => {
        // 首先按是否有活跃订单排序
        if (a.hasActiveOrders && !b.hasActiveOrders) return -1;
        if (!a.hasActiveOrders && b.hasActiveOrders) return 1;
        
        // 如果都有活跃订单，按地板价排序
        if (a.floorPrice && b.floorPrice) {
          return ethers.BigNumber.from(a.floorPrice).lt(ethers.BigNumber.from(b.floorPrice)) ? -1 : 1;
        }
        
        // 如果只有一个有地板价，将有地板价的排在前面
        if (a.floorPrice && !b.floorPrice) return -1;
        if (!b.floorPrice && b.floorPrice) return 1;
        
        // 如果都没有地板价，按名称字母顺序排序
        return a.name.localeCompare(b.name);
      });
    });

    const formatPrice = (price) => {
      if (!price) return '';
      try {
        return ethers.utils.formatEther(price);
      } catch (error) {
        console.error('格式化价格失败:', error, price);
        return '';
      }
    };

    onMounted(fetchCollections);

    return {
      sortedCollections,
      formatPrice,
      loading,
      error,
    };
  },
};
</script>

<style scoped>
.nft-collections {
  margin-top: 20px;
}

.collection-card {
  margin-bottom: 20px;
}

.collection-link {
  text-decoration: none;
  color: inherit;
  display: block;
}

.image {
  width: 100%;
  display: block;
}

.collection-info {
  padding: 14px;
}

.collection-name {
  font-weight: bold;
  color: #333;
}

.bottom {
  margin-top: 13px;
  line-height: 12px;
}

.floor-price {
  color: #666;
}

.loading, .error {
  text-align: center;
  margin-top: 20px;
  font-size: 18px;
}

.error {
  color: red;
}
</style>
