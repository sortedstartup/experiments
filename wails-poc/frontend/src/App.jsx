import { useEffect, useState } from "react";
import logo from "./assets/images/logo-universal.png";
import "./App.css";
import { GetItems,Greet } from "../wailsjs/go/main/App";

function App() {
  const [resultText, setResultText] = useState(
    "Please enter your name below 👇"
  );
  const [name, setName] = useState("");
  const updateName = (e) => setName(e.target.value);
  const updateResultText = (result) => setResultText(result);
  const [item, setItem] = useState(null);

  useEffect(() => {
    GetItems().then((data) => {
        console.log(data.name);
        
      console.log("Received item:", data);
      setItem(data);
    });
  }, []);

  function greet() {
    Greet(name).then(updateResultText);
  }

  return (
    <div id="App">
      <img src={logo} id="logo" alt="logo" />
      <div id="result" className="result">
        {resultText}
      </div>
      <div id="input" className="input-box">
        <input
          id="name"
          className="input"
          onChange={updateName}
          autoComplete="off"
          name="input"
          type="text"
        />
        <button className="btn" onClick={greet}>
          Greet
        </button>
      </div>
      <div>
        <h1>Item from Go:</h1>
        {item && <pre>{JSON.stringify(item, null, 2)}</pre>}
      </div>
    </div>
  );
}

export default App;
