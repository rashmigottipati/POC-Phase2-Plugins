import sys

if __name__ == "__main__":
	print("Executing Python program")
 
	stdin_fileno = sys.stdin
	stdout_fileno = sys.stdout
	stderr_fileno = sys.stderr
 
	sample_input = ['Hello from Python program']
 
	for ip in sample_input:
   		# Print to stdout
		stdout_fileno.write(ip + '\n')
       	try:
			# Adding int to string to raise an exception
        	ip = ip + 100
    	except:
       		stderr_fileno.write('Writing from Python program to stderr: Exception Occurred!\n')

	stdout_fileno.write("Writing args from Python program to stdout: " + sys.argv[1] + ", " + sys.argv[2] + '\n')

	# read from stdin and write to stdout
	for line in sys.stdin:
		sys.stdout.write("Reading stdin data and writing to stdout from python program: " + line + '\n')
