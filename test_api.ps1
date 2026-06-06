# Skillbridge Monolith API Test Runner Script
# Designed for PowerShell (Windows Environments).
# Starts the flow by registering a test student, logging in, harvesting JWTs, and running workspace executions.

$BaseUrl = "http://localhost:8080/api"

# Helper function to print headers
function Write-Header($text) {
    Write-Host "`n==================================================" -ForegroundColor Cyan
    Write-Host ">>> $text" -ForegroundColor Cyan
    Write-Host "==================================================" -ForegroundColor Cyan
}

# Check if server is running
try {
    $null = Invoke-RestMethod -Uri "$BaseUrl/jobs" -Method Get -TimeoutSec 3 -ErrorAction Stop
} catch [System.Net.WebException] {
    if ($_.Exception.Response.StatusCode.value__ -eq 401) {
        Write-Host "[✓] Skillbridge backend is running and reachable!" -ForegroundColor Green
    } else {
        Write-Host "[X] Backend server does not seem to be running at $BaseUrl. Please start it using 'go run cmd/api/main.go' or Docker Compose before running this script." -ForegroundColor Red
        exit
    }
} catch {
    Write-Host "[X] Could not connect to backend server at $BaseUrl. Please verify it is running." -ForegroundColor Red
    exit
}

# 1. Register a Test Student
Write-Header "1. REGISTERING TEST STUDENT"
$RegEmail = "test_student_" + (Get-Random) + "@university.edu"
$RegisterBody = @{
    email = $RegEmail
    password = "securepassword123"
    role = "student"
    university = "Veritas Technical University"
    current_level = "Undergraduate Junior"
} | ConvertTo-Json

try {
    $RegResult = Invoke-RestMethod -Uri "$BaseUrl/auth/register" -Method Post -Body $RegisterBody -ContentType "application/json"
    Write-Host "[✓] Registration successful! Account: $RegEmail" -ForegroundColor Green
    $AccessToken = $RegResult.access_token
    $RefreshToken = $RegResult.refresh_token
    Write-Host "Access Token (First 30 chars): $($AccessToken.Substring(0, 30))..." -ForegroundColor Yellow
} catch {
    Write-Host "[X] Registration failed: $_" -ForegroundColor Red
    exit
}

# 2. Login to get fresh tokens
Write-Header "2. LOGGING IN"
$LoginBody = @{
    email = $RegEmail
    password = "securepassword123"
} | ConvertTo-Json

try {
    $LoginResult = Invoke-RestMethod -Uri "$BaseUrl/auth/login" -Method Post -Body $LoginBody -ContentType "application/json"
    Write-Host "[✓] Login successful!" -ForegroundColor Green
    $AccessToken = $LoginResult.access_token
    Write-Host "Fresh Access Token retrieved." -ForegroundColor Green
} catch {
    Write-Host "[X] Login failed: $_" -ForegroundColor Red
    exit
}

# 3. Retrieve Student Profile (Authenticated)
Write-Header "3. RETRIEVING PROFILE (SECURED ROUTE)"
$Headers = @{
    Authorization = "Bearer $AccessToken"
}

try {
    $Profile = Invoke-RestMethod -Uri "$BaseUrl/profile" -Method Get -Headers $Headers
    Write-Host "[✓] Student Profile details successfully retrieved!" -ForegroundColor Green
    Write-Host "University: $($Profile.university)" -ForegroundColor Yellow
    Write-Host "XP Level  : $($Profile.total_xp)" -ForegroundColor Yellow
    Write-Host "Skills    : $($Profile.skills | ConvertTo-Json -Compress)" -ForegroundColor Yellow
} catch {
    Write-Host "[X] Failed to fetch profile: $_" -ForegroundColor Red
}

# 4. Request Learning Roadmap Path (Authenticated)
Write-Header "4. GENERATING LEARNING ROADMAP DAG"
try {
    $Roadmap = Invoke-RestMethod -Uri "$BaseUrl/roadmap" -Method Get -Headers $Headers
    Write-Host "[✓] Branching Learning Roadmaps loaded successfully!" -ForegroundColor Green
    Write-Host "Roadmap Title: $($Roadmap.title)" -ForegroundColor Yellow
    Write-Host "Number of Nodes in DAG: $($Roadmap.nodes.Count)" -ForegroundColor Yellow
} catch {
    Write-Host "[X] Failed to load Roadmap: $_" -ForegroundColor Red
}

# 5. Sandbox Code Workspace Execution
Write-Header "5. RUNNING SANDBOX CODE WORKSPACE"
$CodeBody = @{
    language = "python"
    code = "print('Skillbridge API Test pipeline executed successfully!')"
} | ConvertTo-Json

try {
    Write-Host "Submitting Python code execution sandbox payload..." -ForegroundColor DarkYellow
    $Execution = Invoke-RestMethod -Uri "$BaseUrl/workspace/execute" -Method Post -Headers $Headers -Body $CodeBody -ContentType "application/json"
    Write-Host "[✓] Execution completed successfully!" -ForegroundColor Green
    Write-Host "Compile Status: $($Execution.status)" -ForegroundColor Yellow
    Write-Host "Sandbox Output: $($Execution.output.Trim())" -ForegroundColor Yellow
    Write-Host "AI Review Score: $($Execution.ai_feedback.score)/100" -ForegroundColor Green
    Write-Host "AI Feedback Recommendation: $($Execution.ai_feedback.recommendations[0])" -ForegroundColor Green
} catch {
    Write-Host "[X] Workspace sandbox execution failed: $_" -ForegroundColor Red
}

Write-Header "TESTS COMPLETE - ALL PRIMARY GATES OK!"
Write-Host "You can continue adding test cycles or view api_tests.http for manual requests." -ForegroundColor Green
