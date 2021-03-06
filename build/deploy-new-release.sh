# sudo systemctl start    https-server.service
# sudo systemctl status   https-server.service   --no-pager

cd /opt/https-server/

ps aux |grep https-server

sudo killall https-server

mv   --force /opt/https-server/https-server-new  /opt/https-server/https-server

sudo setcap cap_net_bind_service=+eip /opt/https-server/https-server

chmod +x /opt/https-server/https-server

# no sudo required
# ./https-server &
#    or
./https-server  >app.log 2>&1 &

less ./app.log
