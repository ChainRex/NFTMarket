<template>
  <div class="nft-collections">
    <h2>NFT 系列</h2>
    <el-row v-if="!loading && !error" :gutter="20">
      <el-col :span="6" v-for="(collection, index) in collections" :key="index">
        <el-card :body-style="{ padding: '0px' }" shadow="hover" class="collection-card">
          <router-link :to="`/collection/${collection.address}`" class="collection-link">
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
import { ref, onMounted } from 'vue';
import { ElSkeleton, ElSkeletonItem } from 'element-plus';
import { ethers } from 'ethers';
import {  getNFTName, getNFTTokenIconURI, getIPFSUrl } from '../utils/nftUtils';
import { initContract, getOrders } from '../utils/contract';

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

        // 初始化合约
        await initContract();
        const rawOrders = await getOrders();
        
        if (!rawOrders || !Array.isArray(rawOrders)) {
          throw new Error('获取到的订单数据无效');
        }

        // 对 NFT 地址进行去重
        const uniqueNFTAddresses = [...new Set(rawOrders.map(order => order.nft))];

        const collectionPromises = uniqueNFTAddresses.map(async (nftAddress) => {
          try {
            const [name, tokenIconURI] = await Promise.all([
              getNFTName(nftAddress),
              getNFTTokenIconURI(nftAddress)
            ]);

            const imageUrl = getIPFSUrl(tokenIconURI);

            const activeOrders = rawOrders.filter(order => 
              order.nft === nftAddress && order.status._hex === '0x00'
            );
            
            let floorPrice = null;
            if (activeOrders.length > 0) {
              floorPrice = activeOrders.reduce((min, order) => {
                const price = ethers.BigNumber.from(order.price);
                return price.lt(min) ? price : min;
              }, ethers.BigNumber.from(activeOrders[0].price));
            }

            return {
              address: nftAddress,
              name,
              imageUrl,
              floorPrice
            };
          } catch (error) {
            console.error('处理 NFT 系列时出错:', error, nftAddress);
            return null;
          }
        });

        const collectionResults = await Promise.all(collectionPromises);
        collections.value = collectionResults.filter(collection => collection !== null);

      } catch (err) {
        console.error('获取 NFT 系列失败:', err);
        error.value = '加载 NFT 系列失败，请稍后重试';
      } finally {
        loading.value = false;
      }
    };

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
      collections,
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