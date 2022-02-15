const msg = {
    phoneNumber: "phoneNumberField.value",
    body:        "bodyField.value",
};

/*
    createObjectStore is synchroneous

*/

var db = {

    dbInner: null,

    init: async () => {
        if (db.dbInner) {
            return Promise.resolve(db.dbInner);
        }
        // init branch
        const fcUpgrade = upgradeDb => {
            const store = upgradeDb.createObjectStore(
                'outbox', 
                {   
                    keyPath: 'id',
                    autoIncrement: true, 
                }
            );
            store.createIndex('date', 'date');
        }
        db.dbInner = await idb.open( 'outbox', 1, fcUpgrade );
        return db.dbInner;
    },


    // const outb = await db.getStore('outbox', 'readwrite');
    getStore: async (name, mode) => {
        const db = await db.init();
        const tx = db.transaction(name, mode);
        return tx.objectStore(name);
    },

    outboxPut: async msg => {
        const db = await db.init();
        const tx = db.transaction('outbox', 'readwrite');
        const tb = tx.objectStore('outbox');  // table
        return await tb.put(msg);
    },

    // register for a sync
    outboxPutAndSync: async msg => {
        try {
            const putState =  await outboxPut(msg);
            console.log(`putState`, {putState});
            if ('serviceWorker' in navigator) {
                const reg = await navigator.serviceWorker.ready;
                return await reg.sync.register('tag-sync-outboxPutAndSync');
            } else {
                return await self.registration.sync.register('tag-sync-outboxPutAndSync');
            }
        } catch (err) {
            console.error(err);
            // form.submit();
        }
    },

}


async function dbExampleExec() {
    console.log(`db example exec start`);
    const outb = await db.getStore('outbox', 'readwrite');
    const val = (await outb.get('counter')) || 0;
    await outb.put(val + 1, 'counter');
    await tx.done;    
    console.log(`db example exec end ${val}`);
    return 1;

}

async function dbExample() {
    console.log(`db example start`);
    let igno = await dbExampleExec();
    console.log(`db example end`);
}