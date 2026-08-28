@ECHO OFF
SET PM2_HOME=C:\Users\LENOVO\.pm2
SET PATH=C:\Program Files\nodejs;%PATH%
cd /d C:\Users\LENOVO\webapp\pusaka-monitor
pm2 resurrect
