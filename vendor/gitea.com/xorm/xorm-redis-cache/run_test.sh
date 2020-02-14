redis-cli FLUSHALL
if [ $? == "0" ];then
    go test -v -run=TestMysqlWithCache
else
    echo "no redis-server running on localhost"
fi
