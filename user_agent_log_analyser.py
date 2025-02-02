import re
import sys
from collections import Counter

def analyze_log(file_path):
    # Regular expression to capture the user agent string
    ua_regex = re.compile(r'Using User-Agent:\s*"([^"]+)"')
    # Regular expression to identify an error message for a failed action
    error_regex = re.compile(r'ERROR\s+Failed to execute action')
    
    # Counter to keep track of how many times each user agent leads to an error
    failed_agents = Counter()
    
    # This variable will temporarily hold the last seen user agent.
    last_user_agent = None

    with open(file_path, 'r') as f:
        for line in f:
            # Check if the line has a user agent log
            ua_match = ua_regex.search(line)
            if ua_match:
                last_user_agent = ua_match.group(1)
                continue  # Go to the next line

            # Check if the line indicates an error
            if error_regex.search(line) and last_user_agent:
                failed_agents[last_user_agent] += 1
                # Reset last_user_agent to avoid counting multiple errors for a single user agent line
                last_user_agent = None

    return failed_agents

def main():
    if len(sys.argv) != 2:
        print("Usage: python analyze_log.py <log_file_path>")
        sys.exit(1)
    
    log_file_path = sys.argv[1]
    results = analyze_log(log_file_path)
    
    # Sort the results by frequency (highest first)
    sorted_results = sorted(results.items(), key=lambda item: item[1], reverse=True)
    
    print("User Agents that resulted in failed actions (ordered by frequency):")
    for user_agent, count in sorted_results:
        print(f"{user_agent} (failed {count} time{'s' if count > 1 else ''})")

if __name__ == "__main__":
    main()
