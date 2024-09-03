import csv
import sys


def copy_csv_lines(source_file, destination_file, num_lines=100):
    with open(source_file, "r") as source:
        reader = csv.reader(source)
        with open(destination_file, "w", newline="") as destination:
            writer = csv.writer(destination)
            for i, row in enumerate(reader):
                if i >= num_lines:
                    break
                writer.writerow(row)


if __name__ == "__main__":
    num_lines = int(sys.argv[3])
    copy_csv_lines(sys.argv[1], sys.argv[2], num_lines)
