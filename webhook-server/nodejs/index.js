const express = require("express");

const app = express();

app.use(express.json());

app.post("/", (req, res) => {
  // Do things here
  console.log(req.body);

  res.sendStatus(201);
});

app.listen(
  8080,
  () => console.log("server listening on http://localhost:8080")
);
