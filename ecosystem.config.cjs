module.exports = {
  apps: [
    {
      name: "pusaka-monitor",
      script: "C:\\Users\\LENOVO\\webapp\\pusaka-monitor\\pusaka-monitor.exe",
      cwd: "C:\\Users\\LENOVO\\webapp\\pusaka-monitor",
      exec_mode: "fork",
      autorestart: true,
      max_restarts: 10,
      restart_delay: 3000,
      env: {
        PORT: "8080",
        GIN_MODE: "release",
      },
      error_file: "C:\\Users\\LENOVO\\webapp\\pusaka-monitor\\logs\\error.log",
      out_file: "C:\\Users\\LENOVO\\webapp\\pusaka-monitor\\logs\\out.log",
      log_file: "C:\\Users\\LENOVO\\webapp\\pusaka-monitor\\logs\\combined.log",
      merge_logs: true,
      time: true,
    },
  ],
};
