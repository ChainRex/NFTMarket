   #!/bin/bash

   # 构建Vue项目
   npm run build

   # 同步文件到Nginx目录
   sudo rsync -av --delete dist/ /var/www/NFTMarket/

   sudo nginx -s reload