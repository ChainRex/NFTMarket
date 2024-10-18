<template>
  <div id="app">
    <NavBar />
    <router-view></router-view>
  </div>
</template>

<script>
import { onMounted } from 'vue';
import { useStore } from 'vuex';
import NavBar from './components/NavBar.vue';

export default {
  name: 'App',
  components: {
    NavBar
  },
  setup() {
    const store = useStore();

    onMounted(async () => {
      await store.dispatch('checkWalletConnection');
    });

    return {};
  }
}
</script>