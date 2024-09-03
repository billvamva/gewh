workers=(50)
queue_sizes=(100)
batch_size=(1000000)

# Input and output paths
input_path="/Users/vasilieiosvamvakas/Documents/projects/gewh/data/weather_data.csv"
output_path="/Users/vasilieiosvamvakas/Documents/projects/gewh/data/output_data.csv"

# Function to run the program and measure execution time
run_test() {
    start_time=$(date +%s.%N)
    ../main  -workers "$1" -queue "$2" -batch "$3" -v true
    end_time=$(date +%s.%N)
    execution_time=$(echo "$end_time - $start_time" | bc)
    echo "Workers: $1, Queue: $2, Batch: $3, Time: $execution_time seconds"
}

# Run tests for all combinations
for w in "${workers[@]}"; do
    for q in "${queue_sizes[@]}"; do
        for b in "${batch_size[@]}"; do
            run_test "$w" "$q" "$b"
        done
    done
done
