
terraregConfig = undefined;
$(document).ready(function () {
    $.ajax({
        type: "GET",
        url: "/v1/terrareg/config",
        success: function (data) {
            terraregConfig = data;
        }
    });
});