import { openDB } from "idb";
// import { isToday } from "date-fns";

export interface TableItem {
  //  id: string;
  national_id: string;
}

export async function dbInsert(
  database: string,
  table: string,
  // id: string,
  item: TableItem
) {
  const db = await openDB(database, 1, {
    upgrade(db) {
      if (!db.objectStoreNames.contains(table)) {
        const createdTable = db.createObjectStore(table);
        createdTable.createIndex("id", "id", { unique: true });
      }
    },
  });
  let stores = await db.getAll(table);
  if (stores.length > 10) {
    const oldest = JSON.parse(stores.shift());
    await db.delete(table, oldest.id);
  }
  const tx = db.transaction(table, "readwrite");
  const store = tx.objectStore(table);
  await store.put(JSON.stringify(item));
  await tx.done;
}

export async function dbGet(database: string, table: string, id: string) {
  const db = await openDB(database, 1, {
    upgrade(db) {
      if (!db.objectStoreNames.contains(table)) {
        const createdTable = db.createObjectStore(table);
        createdTable.createIndex("id", "id", { unique: true });
      }
    },
  });
  return await db.get(table, id);
}

export async function deleteExpiredRecords(database: string, table: string) {
  const db = await openDB(database, 1, {
    upgrade(db) {
      if (!db.objectStoreNames.contains(table)) {
        const createdTable = db.createObjectStore(table);
        createdTable.createIndex("id", "id", { unique: true });
      }
    },
  });
  let stores = await db.getAll(table);

  stores.forEach((each: string) => {
    let parsed: TableItem = JSON.parse(each);
    // if (!isToday(Date.parse(parsed.created_at))) {
    //   db.delete(table, parsed.id);
    // }
  });
}

export async function loadAllRecords(database: string, table: string) {
  const db = await openDB(database, 1, {
    upgrade(db) {
      if (!db.objectStoreNames.contains(table)) {
        const createdTable = db.createObjectStore(table);
        createdTable.createIndex("id", "id", { unique: true });
      }
    },
  });
  return await db.getAll(table);
}

export async function clearTable(database: string, table: string) {
  const db = await openDB(database, 1, {
    upgrade(db) {
      if (!db.objectStoreNames.contains(table)) {
        const createdTable = db.createObjectStore(table);
        createdTable.createIndex("id", "id", { unique: true });
      }
    },
  });
  await db.clear(table);
}
