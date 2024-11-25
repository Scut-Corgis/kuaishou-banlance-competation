
cd B/logs
mkdir 3
cd ..
nohup ./bin/B -log_dir=./logs/3 -hp=10023 -gp=10024 -cpu_num=2 -debug=true -safety_factor=1.5 -max_exec_time=1000000 > ./logs/2/stdout.log 2>&1 &
#./bin/A -log_dir=./logs -hp=44444 -gp=55555


