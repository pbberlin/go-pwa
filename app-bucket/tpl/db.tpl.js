var db = {

    dbInner: null,

    init: async () => {
        if (db.dbInner) {
            return Promise.resolve(db.dbInner);
        }
        // the third arg are four funcs inside {}; we cannot rewrite them as arrow funcs nor factor them out :-(
        db.dbInner = await idb.openDB('db', {{.SchemaVersion}} , {
            upgrade(db, oldVer, newVer, enhTx) {
                // if (db.oldVersion == 0)  db.createObjectStore(...
                if (!db.objectStoreNames.contains('articles')) {
                    const store = db.createObjectStore('articles', { keyPath: 'id', autoIncrement: true });
                    let idx1 = store.createIndex('price_idx', 'price');
                    let idx2 = store.createIndex('date', 'date');
                    console.log(`Vs ${oldVer}--${newVer}: db schema: objectStore created `);
                } else {
                    console.log(`Vs ${oldVer}--${newVer}: db schema: objectStore exists `);
                }

                if (!db.objectStoreNames.contains('table2')) {
                    const store = db.createObjectStore('table2', {});  // no "in-line keys"
                    let idx1 = store.createIndex('idx_name', 'price');
                }


                if (oldVer == 1 && newVer == 2) {
                    // const tx = db.transaction('articles', 'readwrite'); // cannot use; need to use enhTx
                    const store = enhTx.objectStore('articles');
                    let idx2 = store.createIndex('date3', 'date3');
                    console.log(`Vs ${oldVer}--${newVer}: db schema:  index created `);
                }
            },
            blocked() { console.error("blocked") },
            blocked() { console.error("blocking") },
            terminated() { console.error("terminated without db.close()") },
        });
        console.log("db.objectStoreNames", db.objectStoreNames);
        return db.dbInner;
    },


    // const articles = await db.getTableInDB('articles', 'readwrite');
    getTableInDB: async (name, mode) => {
        const db1 = await db.init();
        const tx = db1.transaction(name, mode);
        return tx.objectStore(name);
    },

    /*
    async function getTableInDB(db, name, mode) {
        const tx = db.transaction(name, mode);
        return tx.objectStore(name);
    }

    async function put(db, name, obj) {
        const tbl = await getTableInDB(db, name, "readwrite");
        return await tbl.put(obj);
    }
    */


}


// demo stuff


const msg = {
    phoneNumber: "phoneNumberField.value",
    body:        "bodyField.value",
};

const art1 = {
    title: 'Article 1',
    date: new Date('1819-01-01'),
    date2: new Date('1839-01-01'),
    body: 'content a1',
}

const art2 = {
    title: 'Article 2',
    date: new Date('2019-01-01'),
    date2: new Date('2039-01-01'),
    body: 'content a2',
}



async function doSync() {

    const fcPost = async (msgOrMsgs) => {
        const rawResponse = await fetch('https://localhost/save-json', {
            method: 'POST',
            headers: {
                'Accept': 'application/json',
                'Content-Type': 'application/json'
            },
            body: JSON.stringify(msgOrMsgs),
        });
        const rsp = await rawResponse.json();
        console.log(`save-json response: `, { rsp });
        return rsp;
    }


    // send messages in bulk
    try {
        const outb = await db.getStore('outbox', 'readonly');
        const msgs = await outb.getAll();
        const rsps = await fcPost(msgs);
        console.log(`save-json bulk response: `, { rsps });
    } catch (err) {
        console.error(err);
    }


    // send messages each
    try {
        const outb = await db.getStore('outbox', 'readonly');
        const msgs = await outb.getAll();
        const rsps = await Promise.all(msgs.map(msg => fcPost(msg)));
        console.log(`save-json single response: `, { rsps });
    } catch (err) {
        console.error(err);
    }

}



/* 
    https://github.com/jakearchibald/idb#opendb
    https://javascript.info/indexeddb#object-store

*/




async function dbExample() {

    console.log(`db example start`);

    const dbP = await db.init();
    // Add an article:
    await dbP.add('articles', art1);
    await dbP.add('articles', art2);


    try {
        const tx = dbP.transaction('table2', 'readwrite');
        const tbl = tx.objectStore('table2');
        const val = (await tbl.get('counter')) || 0;
        await tbl.put( val+1, 'counter');
        await tx.done;
        console.log(`a.) atomic counter val is ${val}`);
    } catch (err) {
        console.error(err);
    }

    // short notation for above - based on idb library wrapper
    try {
        const val = (await dbP.get('table2', 'counter')) || 0;
        await dbP.put('table2',  val+1, 'counter');
        console.log(`b.) atomic counter val is ${val}`);
    } catch (err) {
        console.error(err);
    }

    console.log(`db example end`);

}



