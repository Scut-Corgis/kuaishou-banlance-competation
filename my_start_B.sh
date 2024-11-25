
cd B
rm -rf ./logs/*
nohup ./bin/B -log_dir=./logs -hp=10003 -gp=10004 -cpu_num=4 -debug=true > ./logs/stdout.log 2>&1 &
#./bin/A -log_dir=./logs -hp=44444 -gp=55555
