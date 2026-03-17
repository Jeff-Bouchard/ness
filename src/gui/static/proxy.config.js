const PROXY_CONFIG = {
  "/api/*": {
    "target": "http://127.0.0.1:1024",
    "secure": false,
    "logLevel": "debug",
    "bypass": function (req) {
      req.headers["host"] = '127.0.0.1:1024';
      req.headers["referer"] = 'http://127.0.0.1:1024';
      req.headers["origin"] = 'http://127.0.0.1:1024';
    }
  }
};
module.exports = PROXY_CONFIG;
