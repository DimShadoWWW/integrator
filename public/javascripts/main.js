$( "#nav-sidebar" ).load( "menu.html" );

var els = document.getElementsByTagName("a");
for (var i = 0, l = els.length; i < l; i++) {
    var el = els[i];
    if (el.href === '/'+window.location.pathname) {
        el.parentNode.className = 'active';
    }
}
