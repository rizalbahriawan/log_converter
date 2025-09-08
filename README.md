Simple rest API to convert log to timesheet excel file.
Steps:
1. Install go. (v 1.20+)
2. Install Postman Desktop
3. Clone project to local, e.g: C:\log_converter
4. Inside log_converter, open terminal/cmd and execute "go run ." without quote. Note that it will be run on port :8099
5. Open Postman Desktop, create or open your existing workspace.
6. Create new request and type "localhost:8099/convert", set method to POST.
7. Set body request as JSON raw and type request. This is the example
   {
    "username": "username",
    "password": "password",
    "months": [1,2,3],
    "year": 2025,
    "project_name": "your_project_name",
    "randomize_log": {
        "is_random": false,
        "min_duration": 4,
        "max_duration": 6
    }
}
9. Click Send
10. If success, at response tab click options with 3 dots, and click "Save response to file"
11. If error, look for error message.

Notes: To set your duration of work as ranged random numbers, set is_random on randomize_log as true. If is_random is true, you can set minimum (min_duration) and maximum duration (max_duration) of each day, and if maximum duration isn't set, it will be set to your actual duration from the API.
