const { defineConfig } = require('@vue/cli-service')
module.exports = {
  transpileDependencies: true,
  // 其他配置...

  // 如果你需要在构建时注入一些变量
  chainWebpack: config => {
    config.plugin('define').tap(args => {
      const env = args[0]['process.env']
      for (let key in env) {
        // 添加一个检查，避免处理特定的环境变量
        if (key !== 'BASE_URL' && key !== 'PUBLIC_PATH') {
          env[key] = JSON.stringify(env[key])
        }
      }
      return args
    })
  },
  devServer: {
    proxy: {
      '/api': {
        target: 'https://amoy.polygonscan.com',
        changeOrigin: true,
        pathRewrite: {
          '^/api': ''
        }
      }
    }
  }
}

