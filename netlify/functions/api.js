const express = require('express');
const serverless = require('serverless-http');
const menuItems = require('../../menu.json');

const app = express();
const router = express.Router();

// Middleware
app.use(express.json());

// Routes
router.get("/", (req, res) => {
  res.json("Addis Eats Backend");
});

router.get("/menu", (req, res) => {
  res.json({ status: "ok", data: menuItems.menuItems });
});

router.get("/menu/specials", (req, res) => {
    const specials = menuItems.menuItems.filter((item) => item.isSpecial);
    res.json({ status: "ok", data: specials });
  })


// Set the base pathway path
app.use('/.netlify/functions/api', router);

// Export handler instead of using app.listen()
module.exports.handler = serverless(app);
