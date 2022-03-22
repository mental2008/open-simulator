mkdir -p logs
CLUSTER="paib"
CLUSTER_DATE="2022_03_13_11_30_06"
SIMON_BIN="./bin/simon_linux"
for CONFIG in example/scheduler-config; do
  [[ -e "$CONFIG" ]] || break
  echo ${CONFIG}
  for INFLATION in 100 105 110 115 120 125 130 150 200 250 300; do
    CLUSTER_YAML="example/${CLUSTER}_config_${CLUSTER_DATE}_${INFLATION}.yaml"
    for i in {1..19}; do
      echo "logs/${CONFIG}-${INFLATION}-${i}.log"
      time $SIMON_BIN apply --extended-resources "gpu" -f $CLUSTER_YAML --default-scheduler-config "example/scheduler-config/${CONFIG}" > "logs/${CONFIG}-${INFLATION}-${i}.log" & date
    done
    i=20
    echo "logs/${CONFIG}-${INFLATION}-${i}.log"
    time $SIMON_BIN apply --extended-resources "gpu" -f $CLUSTER_YAML --default-scheduler-config "example/scheduler-config/${CONFIG}" > "logs/${CONFIG}-${INFLATION}-${i}.log"
    sleep 5
  done
done

CONFIG="scheduler-config-500x500.yaml"
INFLATION=300
i=20
./bin/simon_linux apply --extended-resources "gpu" -f "example/paib/2022_03_13_11_30_06/paib_config_2022_03_13_11_30_06_${INFLATION}.yaml" --default-scheduler-config "example/scheduler-config/${CONFIG}" > "logs/${CONFIG}-${INFLATION}-${i}.log"