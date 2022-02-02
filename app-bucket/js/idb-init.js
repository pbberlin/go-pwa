const msg = {
    phoneNumber: "phoneNumberField.value",
    body:        "bodyField.value",
};


var store = {
    db: null,

    init: async () => {
        if (store.db) { 
            return Promise.resolve(store.db); 
        }
        // init branch
        const fcUpgrade = (upgradeDb) => {
            upgradeDb.createObjectStore(
                'outbox', { autoIncrement: true, keyPath: 'id' }
            );
        }
        store.db = await idb.open( 'outbox', 1, fcUpgrade );
        return store.db;
    },


    outbox: async mode => {
        const db = await store.init();
        const tx = db.transaction('outbox', mode);
        return tx.objectStore('outbox');
    },

    outboxPut: async msg => {
        const db = await store.init();
        const tx = db.transaction('outbox', 'readwrite');
        return tx.objectStore('outbox').put(msg);
    },

    outboxPutAndSync: async msg => {
        try {
            const db = await store.init();
            const tx = db.transaction('outbox', 'readwrite');
            const putState =  await tx.objectStore('outbox').put(msg);
            console.error(`putState`, {putState});
            if (false) {
                phoneNumberField.value = '';
            }
            return reg.sync.register('tag-sync-outboxPutAndSync');              
        } catch (err) {
            console.error(err);
            form.submit();
        }
        
    },

}
