<template>
  <div class="collection-detail">
    <el-page-header @back="goBack">
      <template #icon>
        <el-icon class="page-header-icon"><Back /></el-icon>
      </template>
      <template #content>
        <el-skeleton :loading="loading" animated>
          <template #template>
            <el-skeleton-item variant="text" style="width: 150px;" />
          </template>
          <template #default>
            {{ collection.Name || 'NFT 系列' }}
          </template>
        </el-skeleton>
      </template>
    </el-page-header>
    
    <!-- 集合头部信息 -->
    <el-skeleton :loading="loading" animated>
      <template #template>
        <el-row class="collection-header">
          <el-col :span="4">
            <el-skeleton-item variant="circle" style="width: 100px; height: 100px;" />
          </el-col>
          <el-col :span="20">
            <el-skeleton-item variant="h1" style="width: 50%;" />
            <el-skeleton-item variant="text" style="width: 60%;" />
            <el-skeleton-item variant="text" style="width: 40%;" />
          </el-col>
        </el-row>
      </template>
      <template #default>
        <el-row class="collection-header">
          <el-col :span="4">
            <el-avatar :size="100" :src="collection.TokenIconURI"></el-avatar>
          </el-col>
          <el-col :span="20">
            <h1>
              <a :href="getExplorerUrl('token', collection.ContractAddress)" target="_blank" rel="noopener noreferrer">
                {{ collection.Name }}
              </a>
            </h1>
            <p class="contract-address">
              合约地址: 
              <a :href="getExplorerUrl('address', collection.ContractAddress)" target="_blank" rel="noopener noreferrer">
                {{ collection.ContractAddress }}
              </a>
              <el-button class="copy-button" @click="copyAddress(collection.ContractAddress)" type="text">
                <el-icon><DocumentCopy /></el-icon>
              </el-button>
            </p>
            <p>代号: {{ collection.Symbol }}</p>
          </el-col>
        </el-row>
      </template>
    </el-skeleton>

    <!-- NFT 列表 -->
    <el-row :gutter="20">
      <template v-if="loading">
        <el-col :span="6" v-for="i in 8" :key="i">
          <el-card :body-style="{ padding: '0px' }" shadow="hover">
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
      </template>
      
      <template v-else>
        <el-col :span="6" v-for="nft in nfts" :key="nft.TokenID">
          <router-link :to="`/nft/${collection.ContractAddress}/${nft.TokenID}`" class="nft-link">
            <el-card :body-style="{ padding: '0px' }" shadow="hover">
              <img :src="nft.Image" class="image" :alt="nft.Name">
              <div style="padding: 14px;">
                <span>{{ collection.Name }} #{{ nft.TokenID }} - {{ nft.Name }}</span>
                <div class="bottom">
                  <span v-if="nft.price && nft.orderStatus === 0" class="price">{{ formatPrice(nft.price) }} {{ nft.tokenSymbol }}</span>
                  <span v-else class="not-for-sale">暂无出售</span>
                </div>
              </div>
            </el-card>
          </router-link>
        </el-col>
      </template>
    </el-row>
  </div>
</template>

<script>
import { ref, onMounted, watch } from 'vue';
import { useRoute, useRouter } from 'vue-router';
import { ethers } from 'ethers';
import { ElSkeleton, ElSkeletonItem, ElMessage } from 'element-plus';
import { Back, DocumentCopy } from '@element-plus/icons-vue';
import axios from 'axios';
import { getTokenInfo } from '../utils/nftUtils'; // 确保导入这个函数

const API_BASE_URL = 'http://121.196.204.174:8081/api';

export default {
  components: {
    ElSkeleton,
    ElSkeletonItem,
    Back,
    DocumentCopy,
  },
  setup() {
    const route = useRoute();
    const router = useRouter();
    const collection = ref({});
    const nfts = ref([]);
    const loading = ref(true);

    const fetchCollectionDetails = async () => {
      try {
        loading.value = true;
        const collectionAddress = route.params.address;
        
        // 从后端获取集合详情
        const response = await axios.get(`${API_BASE_URL}/nft/${collectionAddress}`);
        const data = response.data;

        collection.value = data.collection;
        nfts.value = data.nfts;

        // 获取每个 NFT 的订单信息
        await Promise.all(nfts.value.map(async (nft) => {
          try {
            const orderResponse = await axios.get(`${API_BASE_URL}/order/${collectionAddress}/${nft.TokenID}`);
            const orderData = orderResponse.data;
            if (orderData && !orderData.error) {
              const tokenInfo = await getTokenInfo(orderData.TokenAddress);
              nft.price = orderData.Price;
              nft.tokenSymbol = tokenInfo.symbol;
              nft.orderStatus = orderData.Status;
            } else {
              nft.price = null;
              nft.tokenSymbol = null;
              nft.orderStatus = null;
            }
          } catch (error) {
            if (error.response && error.response.data && error.response.data.error === "订单未找到") {
              console.log(`NFT #${nft.TokenID} 当前没有出售`);
              nft.price = null;
              nft.tokenSymbol = null;
              nft.orderStatus = null;
            } else {
              console.error(`获取 NFT #${nft.TokenID} 订单信息失败:`, error);
            }
          }
        }));

      } catch (error) {
        console.error('获取集合详情失败:', error);
        ElMessage.error('获取集合详情失败: ' + error.message);
      } finally {
        loading.value = false;
      }
    };

    const formatPrice = (price) => {
      if (!price) return 'N/A';
      try {
        return ethers.utils.formatEther(price);
      } catch (error) {
        console.error('格式化价格失败:', error, price);
        return 'Error';
      }
    };

    const goBack = () => {
      router.back();
    };

    const getExplorerUrl = (type, address) => {
      const baseUrl = 'https://amoy.polygonscan.com';
      switch (type) {
        case 'token':
          return `${baseUrl}/token/${address}`;
        case 'address':
          return `${baseUrl}/address/${address}`;
        default:
          return baseUrl;
      }
    };

    const copyAddress = (address) => {
      navigator.clipboard.writeText(address).then(() => {
        ElMessage.success('地址已复制到剪贴板');
      }).catch(err => {
        console.error('复制失败:', err);
        ElMessage.error('复制地址失败');
      });
    };

    onMounted(fetchCollectionDetails);

    watch(
      () => route.params.address,
      (newAddress, oldAddress) => {
        if (newAddress !== oldAddress) {
          fetchCollectionDetails();
        }
      }
    );

    return {
      collection,
      nfts,
      loading,
      formatPrice,
      goBack,
      getExplorerUrl,
      copyAddress
    };
  }
};
</script>

<style scoped>
.collection-detail {
  max-width: 1200px;
  margin: 0 auto;
  padding: 20px;
}

.collection-header {
  margin: 20px 0;
}

.image {
  width: 100%;
  display: block;
}

.bottom {
  margin-top: 13px;
  line-height: 12px;
}

.el-card {
  margin-bottom: 20px;
}

.nft-link {
  text-decoration: none;
  color: inherit;
}

a {
  color: #409EFF;
  text-decoration: none;
}

a:hover {
  text-decoration: none; /* 移除悬浮时的下划线 */
}

.page-header-icon {
  margin-right: 8px;
}

.contract-address {
  display: flex;
  align-items: center;
}

.copy-button {
  margin-left: 8px;
  padding: 2px;
}

.copy-button .el-icon {
  font-size: 16px;
}
</style>
