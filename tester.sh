URL="http://localhost:8080"

echo "started"


NUM_REQUESTS=1000


start_time=$(date +%s)

for ((i=1; i<=NUM_REQUESTS; i++))
do
  curl -s -o /dev/null -w "%{http_code}\n" $URL &
done


wait

end_time=$(date +%s)
elapsed_time=$((end_time - start_time))

echo "Sent $NUM_REQUESTS requests to $URL"
echo "Time taken: $elapsed_time seconds"