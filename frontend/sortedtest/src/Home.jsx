import React, { useState } from "react";

const items = [
  { id: 1, name: "Apple", info: "A sweet red fruit" },
  { id: 2, name: "Banana", info: "A long yellow fruit" },
  { id: 3, name: "Carrot", info: "An orange root vegetable" },
];

export default function Home() {
  const [search, setSearch] = useState("");
  const filtered = items.filter(item => item.name.toLowerCase().includes(search.toLowerCase()));

  return (
    <div>
      <h2>Home Page</h2>
      <input
        type="text"
        placeholder="Search..."
        value={search}
        onChange={e => setSearch(e.target.value)}
        style={{ marginBottom: 10 }}
      />
      <ul>
        {filtered.map(item => (
          <li key={item.id}>
            <strong>{item.name}</strong>: {item.info}
          </li>
        ))}
      </ul>
    </div>
  );
}
