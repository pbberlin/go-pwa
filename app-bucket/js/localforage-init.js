// https://localforage.github.io/localForage/


const dbName = 'db1';
const store1 = 'table1';
const desc1  = 'user profile';
const store2 = 'table2';
const desc2  = 'user config';

// dropping WEBSQL instance never removed the version 1.0
//   neither calling dropInstance() explicitly for version 1.0
//       nor implicitly without any version helps
const vs     =  1.0;


const cfg1 = {
    name:         dbName, // different database name, instead of localforage
    storeName:    store1, // alphanumeric with underscores
    description:  desc1 ,
    driver: [                         // different driver order
        localforage.WEBSQL,
        localforage.INDEXEDDB,
        localforage.LOCALSTORAGE,
    ],
    version: vs,
    size: 4980736, // database size in bytes. WebSQL-only
}

const fcInitDB = async () => {    

    // for multiple 'tables' we use localforage.createInstance() instead of localforage.config();
    //   side-effect: an IndexDB - localforage - keyvaluepairs is still created
    var table1 = await localforage.createInstance(cfg1);
    console.log(`created store ${store1} - in database ${dbName}`);

    let cfg2 = cfg1;

    cfg2.storeName = store2;
    cfg2.description = desc2;
    var table2 = await localforage.createInstance(cfg2);
    console.log(`created store ${store2} - in database ${dbName}`);

    // stores only become visible in chrome dev tools
    // after the first insert

    // return;

    const userProfile = {name: "Max", surname: "Factor"}
    const userConfig  = {sort: "price", direction: "ascending"}


    table1.setItem("key1", `val-${vs}`   ).then(  () => console.log("key1 stored in table1") );
    table1.setItem("profile", userProfile).then(  () => console.log("profile stored in table1") );
    const fcKeys = await table1.keys;
    // console.log(`keys of table1 are ${fcKeys}`);
    const keysRes = await fcKeys();
    console.log(`table1 keys: ${keysRes}`);

    

    table2.setItem("key2", `val-${vs}`   ).then(  () => console.log("key2 stored in table2") );
    table2.removeItem("key2").then(  () => console.log("key2 deleted") );
    table2.setItem("config", userConfig  ).then(  () => console.log("config stored in table2") );
    // table2.clear();

 
    console.log('  iter start');
    table1.iterate(  (val, key, cnt) => console.log(`   itr  ${cnt} - ${key}  ${val}`) )
    .then(function () {
        console.log('  iter stop');
    }).catch(function (err) {
        console.log(`iter error: ${err}`);
    });

    try {
        const val = await table1.getItem('profile');
        console.log(`val for key 'profile' is  ${val}`);
    } catch (err) {
        console.log(`could not getItem(...): ${err}`);
    }

}




try {

    localforage.dropInstance({
        name:        dbName,
        storeName:   store1,
        // version:     vs,  // we want to delete all versions
    })
    .then(   () => console.log(`dropped store ${store1} - from database ${dbName}`)   )
    ;

    localforage.dropInstance({
        name:        dbName,
        // version:     vs,   // we want to delete all versions
    })
    .then(   () => console.log(`dropped store *      - from database ${dbName}`)   )
    .then(   fcInitDB   ) 
    ;

} catch (err) {
    console.log(`db might not yet exist; ${err}`);
}


