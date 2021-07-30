{
  "diskType": "hdd",
  "pool": "hdd_write_pool",
  "volume": "image",
  "runtime": 120,
  "ipAddr": ["10.0.20.27", "10.0.20.28", "10.0.20.29"],
  "osdNum": [0,1],
  "opType": ["write"],
  "blockSize": ["1k", "2k", "4k", "8k", "16k", "32k", "64k", "128k", "256k", "512k", "1M", "2M", "4M", "8M"],
  "ioDepth": [1, 2, 4, 8, 16, 32, 64, 128, 512],
  "recoveryLimit": [0, 79, 158, 316, 500, 700, 999, 1250, 1500, 2000, 4000]
}

ceph osd pool create bd_pool 128

ceph osd pool set hdd_1_pool size 2
ceph osd pool set hdd_2_pool size 2
ceph osd pool set hdd_3_pool size 2
ceph osd pool set hdd_4_pool size 2
ceph osd pool set hdd_5_pool size 2
ceph osd pool set ssd_1_pool size 2
ceph osd pool set ssd_2_pool size 2

ceph osd pool set hdd_1_pool pg_num 128
ceph osd pool set hdd_2_pool pg_num 128
ceph osd pool set hdd_3_pool pg_num 128
ceph osd pool set hdd_4_pool pg_num 128
ceph osd pool set hdd_5_pool pg_num 128
ceph osd pool set ssd_1_pool pg_num 128
ceph osd pool set ssd_2_pool pg_num 128

rbd create -p hdd_1_pool image --size 102400
rbd create -p hdd_2_pool image --size 102400
rbd create -p hdd_3_pool image --size 102400
rbd create -p hdd_4_pool image --size 102400
rbd create -p hdd_5_pool image --size 102400
rbd create -p ssd_1_pool image --size 102400
rbd create -p ssd_2_pool image --size 102400
rbd create -p hdd_rb1_pool image --size 102400
rbd create -p hdd_rb2_pool image --size 102400
rbd create -p hdd_rb3_pool image --size 102400
rbd create -p hdd_rb4_pool image --size 102400
rbd create -p hdd_rb5_pool image --size 102400
rbd create -p ssd_rb2_pool image --size 102400


nohup rbd -p hdd_1_pool bench image --io-type write --io-size 4M --io-threads 32 --io-total 100G --io-pattern seq &
nohup rbd -p hdd_2_pool bench image --io-type write --io-size 4M --io-threads 32 --io-total 100G --io-pattern seq &
nohup rbd -p hdd_3_pool bench image --io-type write --io-size 4M --io-threads 32 --io-total 100G --io-pattern seq &
nohup rbd -p hdd_4_pool bench image --io-type write --io-size 4M --io-threads 32 --io-total 100G --io-pattern seq &
nohup rbd -p hdd_5_pool bench image --io-type write --io-size 4M --io-threads 32 --io-total 100G --io-pattern seq &
nohup rbd -p ssd_1_pool bench image --io-type write --io-size 4M --io-threads 32 --io-total 100G --io-pattern seq &
nohup rbd -p ssd_2_pool bench image --io-type write --io-size 4M --io-threads 32 --io-total 100G --io-pattern seq &

nohup rbd -p hdd_rb1_pool bench image --io-type write --io-size 4M --io-threads 32 --io-total 100G --io-pattern seq &
nohup rbd -p hdd_rb2_pool bench image --io-type write --io-size 4M --io-threads 32 --io-total 100G --io-pattern seq &
nohup rbd -p hdd_rb3_pool bench image --io-type write --io-size 4M --io-threads 32 --io-total 100G --io-pattern seq &
nohup rbd -p hdd_rb4_pool bench image --io-type write --io-size 4M --io-threads 32 --io-total 100G --io-pattern seq &
nohup rbd -p hdd_rb5_pool bench image --io-type write --io-size 4M --io-threads 32 --io-total 100G --io-pattern seq &
nohup rbd -p ssd_rb2_pool bench image --io-type write --io-size 4M --io-threads 32 --io-total 100G --io-pattern seq &




nohup /home/liucx/hdd/get_all_cost_write/get_all_cost       /home/liucx/hdd/get_all_cost_write/config.json >        /home/liucx/hdd/get_all_cost_write/cost.log &
nohup /home/liucx/hdd/get_all_cost_read/get_all_cost        /home/liucx/hdd/get_all_cost_read/config.json >         /home/liucx/hdd/get_all_cost_read/cost.log &
nohup /home/liucx/hdd/get_all_cost_randwrite/get_all_cost   /home/liucx/hdd/get_all_cost_randwrite/config.json >    /home/liucx/hdd/get_all_cost_randwrite/cost.log &
nohup /home/liucx/hdd/get_all_cost_randread/get_all_cost    /home/liucx/hdd/get_all_cost_randread/config.json >     /home/liucx/hdd/get_all_cost_randread/cost.log &
nohup /home/liucx/hdd/get_all_cost_rw/get_all_cost          /home/liucx/hdd/get_all_cost_rw/config.json >           /home/liucx/hdd/get_all_cost_rw/cost.log &
nohup /home/liucx/hdd/get_all_cost_randrw/get_all_cost      /home/liucx/hdd/get_all_cost_randrw/config.json >       /home/liucx/hdd/get_all_cost_randrw/cost.log &
nohup /home/liucx/ssd/get_all_cost_write/get_all_cost       /home/liucx/ssd/get_all_cost_write/config.json >        /home/liucx/ssd/get_all_cost_write/cost.log &
nohup /home/liucx/ssd/get_all_cost_read/get_all_cost        /home/liucx/ssd/get_all_cost_read/config.json >         /home/liucx/ssd/get_all_cost_read/cost.log &
nohup /home/liucx/ssd/get_all_cost_randwrite/get_all_cost   /home/liucx/ssd/get_all_cost_randwrite/config.json >    /home/liucx/ssd/get_all_cost_randwrite/cost.log &


cat /home/liucx/hdd/get_all_cost_write/cost.log  |grep "fioConfig Result"
cat /home/liucx/hdd/get_all_cost_read/cost.log  |grep "fioConfig Result"
cat /home/liucx/hdd/get_all_cost_randwrite/cost.log  |grep "fioConfig Result"
cat /home/liucx/hdd/get_all_cost_randread/cost.log  |grep "fioConfig Result"
cat /home/liucx/hdd/get_all_cost_rw/cost.log  |grep "fioConfig Result"
cat /home/liucx/hdd/get_all_cost_randrw/cost.log  |grep "fioConfig Result"
cat /home/liucx/ssd/get_all_cost_read/cost.log  |grep "fioConfig Result"
cat /home/liucx/ssd/get_all_cost_write/cost.log  |grep "fioConfig Result"
cat /home/liucx/ssd/get_all_cost_randwrite/cost.log  |grep "fioConfig Result"


nohup /home/liucx/get_all_cost /home/liucx/ssd_read/config.json > /home/liucx/ssd_read/config.json.log &
nohup /home/liucx/get_all_cost /home/liucx/hdd_randrw/config1.json > /home/liucx/hdd_randrw/config1.json.log &
nohup /home/liucx/get_all_cost /home/liucx/hdd_randrw/config2.json > /home/liucx/hdd_randrw/config2.json.log &
nohup /home/liucx/get_all_cost /home/liucx/hdd_rw/config1.json > /home/liucx/hdd_rw/config1.json.log &
nohup /home/liucx/get_all_cost /home/liucx/hdd_rw/config2.json > /home/liucx/hdd_rw/config2.json.log &
