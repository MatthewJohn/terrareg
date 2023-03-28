function getLocalStorageValue(key, defaultValue, type) {
    let localStorageValue = localStorage.getItem(key);
    let value = localStorageValue === null ? defaultValue : localStorageValue;

    if (type === Boolean) {
        if (value == "false") {
            value = false;
        } else if (value == "true") {
            value = true;
        }
    }
    return value;
}

function getUserPreferences() {
    return {
        'show-beta-versions': getLocalStorageValue('show-beta-versions', false, Boolean),
        'show-unpublished-versions': getLocalStorageValue('show-unpublished-versions', false, Boolean),
        'theme': $.cookie("theme") || 'default'
    }
}

function userPreferencesModalShow() {
    let currentPreferences = getUserPreferences();
    $('#user-preferences-show-beta').prop('checked', currentPreferences["show-beta-versions"]);
    $('#user-preferences-show-unpublished').prop('checked', currentPreferences["show-unpublished-versions"]);
    $('#user-preferences-theme').val(currentPreferences["theme"]);

    $('#user-preferences-modal')[0].classList.add('is-active');
}

function userPreferencesModalSave() {
    let newTheme = $('#user-preferences-theme').val();
    let themeHasChanged = newTheme !== getUserPreferences().theme;

    localStorage.setItem('show-beta-versions', $('#user-preferences-show-beta').is(':checked'));
    localStorage.setItem('show-unpublished-versions', $('#user-preferences-show-unpublished').is(':checked'));
    localStorage.setItem('theme', $('#user-preferences-theme').val());
    $.cookie("theme", newTheme, {path: "/"});
    userPreferencesModalClose();

    if (themeHasChanged) {
        location.reload();
    }
}

function userPreferencesModalClose() {
    $('#user-preferences-modal')[0].classList.remove('is-active');
}