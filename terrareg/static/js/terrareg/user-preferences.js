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
        'theme': getLocalStorageValue('theme', 'default', String)
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
    localStorage.setItem('show-beta-versions', $('#user-preferences-show-beta').is(':checked'));
    localStorage.setItem('show-unpublished-versions', $('#user-preferences-show-unpublished').is(':checked'));
    localStorage.setItem('theme', $('#user-preferences-theme').val());
    userPreferencesModalClose();
}

function userPreferencesModalClose() {
    $('#user-preferences-modal')[0].classList.remove('is-active');
}

function setCurrentTheme() {
    // Load theme CSS
    let userPreferences = getUserPreferences();
    let url = `/static/css/bulma/${userPreferences['theme']}/bulmaswatch.min.css`;
    if (userPreferences['theme'] == 'default') {
      url = '/static/css/bulma/bulma-0.9.3.min.css';
    }
    var head  = document.getElementsByTagName('head')[0];
    var link  = document.createElement('link');
    link.id   = 'bulma';
    link.rel  = 'stylesheet';
    link.type = 'text/css';
    link.href = url;
    link.media = 'all';
    head.appendChild(link);
}

setCurrentTheme();
