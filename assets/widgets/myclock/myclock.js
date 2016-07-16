var bind = function(fn, me){ return function(){ return fn.apply(me, arguments); }; },
  extend = function(child, parent) { for (var key in parent) { if (hasProp.call(parent, key)) child[key] = parent[key]; } function ctor() { this.constructor = child; } ctor.prototype = parent.prototype; child.prototype = new ctor(); child.__super__ = parent.prototype; return child; },
  hasProp = {}.hasOwnProperty;

Dashing.myclock = (function(superClass) {
  var hourL, minuteL, pi, radius, secondL, size;

  extend(myclock, superClass);

  function myclock() {
    this.drawTime = bind(this.drawTime, this);
    return myclock.__super__.constructor.apply(this, arguments);
  }

  size = 286;

  pi = Math.PI;

  radius = size / 2;

  hourL = 0.5 * radius;

  minuteL = 0.75 * radius;

  secondL = 0.9 * radius;

  myclock.prototype.ready = function() {
    return setInterval(this.drawTime, 1000);
  };

  myclock.prototype.drawTime = function() {
    var alpha, beta, canvas, content, gamma, h, m, n, s, theta, today, x, y;
    $(this.node).children("canvas").remove();
    today = new Date();
    h = today.getHours();
    m = today.getMinutes();
    s = today.getSeconds();
    canvas = document.createElement('canvas');
    canvas.setAttribute('height', size);
    canvas.setAttribute('width', size);
    $(this.node).append(canvas);
    content = canvas.getContext('2d');
    content.translate(radius, radius);
    content.beginPath();
    content.arc(0, 0, radius, 0, 2 * pi, false);
    content.closePath();
    content.fillStyle = '#cccccc';
    content.fill();
    content.font = '12px Arial';
    content.fillStyle = '#000';
    content.textAlign = 'center';
    content.textBaseline = 'middle';
    n = 1;
    while (n <= 12) {
      theta = (n - 3) * pi * 2 / 12;
      x = radius * 0.9 * Math.cos(theta);
      y = radius * 0.9 * Math.sin(theta);
      content.fillText(n, x, y);
      n++;
    }
    content.lineWidth = '4';
    content.strokeStyle = 'rgb(255,0,0)';
    content.lineCap = 'round';
    content.beginPath();
    content.moveTo(0, 0);
    alpha = (s - 15) * pi * 2 / 60;
    content.lineTo(secondL * Math.cos(alpha), secondL * Math.sin(alpha));
    content.stroke();
    content.strokeStyle = 'rgb(0,0,0)';
    content.lineWidth = '6';
    content.lineCap = 'round';
    content.beginPath();
    content.moveTo(0, 0);
    beta = (m - 15) * pi * 2 / 60;
    content.lineTo(minuteL * Math.cos(beta), minuteL * Math.sin(beta));
    content.stroke();
    content.lineWidth = '8';
    content.lineCap = 'round';
    content.beginPath();
    content.moveTo(0, 0);
    gamma = ((h - 3) + (m / 60)) * pi * 2 / 12;
    content.lineTo(hourL * Math.cos(gamma), hourL * Math.sin(gamma));
    return content.stroke();
  };

  return myclock;

})(Dashing.Widget);
