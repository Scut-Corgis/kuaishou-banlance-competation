
cd B/logs
mkdir 2
cd ..
nohup ./bin/B -log_dir=./logs/2 -hp=10013 -gp=10014 -cpu_num=2 -debug=false -safety_factor=1.5 -max_exec_time=1000000 > ./logs/2/stdout.log 2>&1 &
#./bin/A -log_dir=./logs -hp=44444 -gp=55555


