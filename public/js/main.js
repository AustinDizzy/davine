$(function(){

    $('ul.nav a[href="'+ window.location +'"]').parent().addClass('active');
    
    $('ul.nav a').filter(function() {
        return this.href == window.location;
    }).parent().addClass('active');

    $('.formatInt').each(function(i) {
        val = $('.formatInt')[i].innerText;
        str = numeral(parseInt(val));
        $('.formatInt')[i].innerText = str.format($(this).data("format"));
    });

    var menu = $('#main-menu');
    var origOffsetY = menu.offset().top;

    function scroll() {
        if ($(window).scrollTop() >= origOffsetY) {
            $('#main-menu').addClass('navbar-fixed-top');
            $('.main').css({"padding-top": "72px"});
            $('.navbar-brand').fadeIn('fast');
        } else {
            $('#main-menu').removeClass('navbar-fixed-top');
            $('.main').removeAttr('style');
            $('.navbar-brand').fadeOut('fast');
        }
    }

    document.onscroll = scroll;
});