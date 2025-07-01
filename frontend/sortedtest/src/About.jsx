import React, { useEffect, useState } from "react";
import { faro, LogLevel } from "@grafana/faro-web-sdk";

const people = [
  { id: 1, name: "Alice", info: "Frontend Developer" },
  { id: 2, name: "Bob", info: "Backend Developer" },
  { id: 3, name: "Charlie", info: "DevOps Engineer" },
];

export default function About() {
  const [search, setSearch] = useState("");
  const filtered = people.filter((person) =>
    person.name.toLowerCase().includes(search.toLowerCase())
  );
  useEffect(() => {
    faro.api.pushLog(["User clicked add to cart"]);
    console.warn("in about page");
    console.info("sanskar page");

    const error = new Error("I am supposed to fail");
    faro.api.pushError(error);
    faro.api.pushEvent("pushevent");
    faro.api.pushMeasurement()
  }, []);

  return (
    <div>
      <h2>About Page</h2>
      <input
        type="text"
        placeholder="Search people..."
        value={search}
        onChange={(e) => setSearch(e.target.value)}
        style={{ marginBottom: 10 }}
      />
      <ul>
        {filtered.map((person) => (
          <li key={person.id}>
            <strong>{person.name}</strong>: {person.info}
          </li>
        ))}
      </ul>
    </div>
  );
}
