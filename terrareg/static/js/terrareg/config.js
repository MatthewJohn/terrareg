
terraregConfig = undefined;

async function getConfig() {
    return new Promise((resolve, reject) => {

        // Check if terrareg config has already been obtained
        if (terraregConfig !== undefined) {
            resolve(terraregConfig);
        }

        // If not, perform request
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
