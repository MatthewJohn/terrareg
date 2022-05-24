
terraregConfigPromiseSingleton = undefined;

async function getConfig() {
    // Create promise if it hasn't already been defined
    if (terraregConfigPromiseSingleton === undefined) {
        terraregConfigPromiseSingleton = new Promise((resolve, reject) => {
            // Perform request to obtain the config
            $.ajax({
                type: "GET",
                url: "/v1/terrareg/config",
                success: function (data) {
                    terraregConfig = data;
                    resolve(terraregConfig);
                }
            });
        });
    }
    return terraregConfigPromiseSingleton;
}

