cd mongo
# 解压数据库文件
tar xvf mongodb-macos-x86_64-4.4.5.tgz
# 拷贝bin文件夹到当前路径
cp -R mongodb-macos-x86_64-4.4.5/bin .
# 拷贝文成后删除原始解压文件夹
rm -f -r mongodb-macos-x86_64-4.4.5
# 创建存放数据库的文件夹
mkdir -p data/db

# 不设置访问权限启动mongodb
./bin/mongod --port 27017 --dbpath data/db

# lsof -i tcp:27017
# kill

# 链接到数据库
#./bin/mongo --port 27017
#
## 创建administrator账户
#use admin
#db.createUser(
#  {
#    user: "xausee",
#    pwd: "xausee",
#    roles: [
#        {
#            role: "userAdminAnyDatabase",
#            db: "admin"
#        },
#        {
#            role: "root",
#            db: "admin"
#        }]
#  }
#)
## 关掉数据库
#db.shutdownServer()
#exit
#
## 以设置权限和指定外网访问的方式重新启动数据库
#./bin/mongod --bind_ip 0.0.0.0 --port 27017 --dbpath data/db --auth
#
## 以管理员的方式访问数据库
#./bin/mongo --host localhost --port 27017 -u "xausee" -p "xausee" --authenticationDatabase "admin"
#
## 创建数据库CaoMang的普通用户:
#use CaoMang
#db.createUser({
#    user: "caomang",
#    pwd: "caomang",
#    roles: [
#        {
#            role: "readWrite",
#            db: "CaoMang"
#        }
#    ]
#})
#
## 创建测试环境数据库CaoMangUAT的普通用户:
#use CaoMangUAT
#db.createUser({
#    user: "caomanguat",
#    pwd: "caomanguat",
#    roles: [
#        {
#            role: "readWrite",
#            db: "CaoMangUAT"
#        }
#    ]
#})
#
#
## 创建数据库SouYun的普通用户:
#use SouYun
#db.createUser({
#    user: "souyun",
#    pwd: "souyun",
#    roles: [
#        {
#            role: "readWrite",
#            db: "SouYun"
#        }
#    ]
#})
