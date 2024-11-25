cd A
rm -rf ./logs/*
nohup ./bin/A -log_dir=./logs -hp=10001 -cpu_num=2 -debug=true -conn_per_pool_num=1 > ./logs/stdout.log 2>&1 &
#./bin/A -log_dir=./logs -hp=44444 -gp=55555
