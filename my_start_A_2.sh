cd A/logs
mkdir 2
cd ..
nohup ./bin/A -log_dir=./logs/2 -hp=10101 -cpu_num=2 -debug=true -conn_per_pool_num=1 > ./logs/2/stdout.log 2>&1 &
#./bin/A -log_dir=./logs -hp=44444 -gp=55555




