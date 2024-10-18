<template>
  <div class="home-page">
    <NFTCollections />
  </div>
</template>

<script>
import NFTCollections from '../components/NFTCollections.vue';

export default {
  name: 'HomePage',
  components: {
    NFTCollections,
  },
};
</script>

<style scoped>
.home-page {
  max-width: 1200px;
  margin: 0 auto;
  padding: 20px;
}
</style>