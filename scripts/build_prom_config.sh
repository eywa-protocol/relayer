cat > ../.data/prometheus.yaml <<EOF
global:
  scrape_interval:     15s
  evaluation_interval: 15s
  external_labels:
    monitor: 'node-monitor'
rule_files:
scrape_configs:
EOF

set -o allexport
  source ../.env.prom
set +o allexport

if [ "$PROM_LISTEN_PORT" != "" ];then
   cat >> ../.data/prometheus.yaml <<EOF
  - job_name: 'bridge-nodes'
    scrape_interval: 5s
    static_configs:
      - targets:
EOF
   for i in {11..17} ; do
     cat >> ../.data/prometheus.yaml <<EOF
          - '172.20.32.$i:$PROM_LISTEN_PORT'
EOF
   done
fi