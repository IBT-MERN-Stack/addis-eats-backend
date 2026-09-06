const express = require('express');
const cors = require('cors');
const menuItems = require('./menu.json');

const app = express();
const PORT = 3000;



// cors setup
app.use(cors());

// parse json bodies
app.use(express.json());

// Routes
app.get("/", (req, res) => {
  res.json("Addis Eats Backend");
});

app.get("/menu", (req, res) => {
  res.json({ status: "ok", data: menuItems.menuItems });
});

app.get("/menu/specials", (req, res) => {
    const specials = menuItems.menuItems.filter((item) => item.isSpecial);
    res.json({ status: "ok", data: specials });
  })





// Start server
app.listen(PORT, () => {
  console.log(`Server running on http://localhost:${PORT}`);
});
