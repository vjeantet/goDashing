var DashboardSwitcher, DashboardSwitcherControls, WidgetSwitcher;
var __bind = function(fn, me){ return function(){ return fn.apply(me, arguments); }; };
DashboardSwitcher = (function() {
  function DashboardSwitcher() {
    var name, names;
    this.dashboardNames = [];
    names = $('[data-switcher-dashboards]').first().attr('data-switcher-dashboards') || '';
    if (names.length > 1) {
      this.dashboardNames = (function() {
        var _i, _len, _ref, _results;
        _ref = names.split(/[ ,]+/).filter(Boolean);
        _results = [];
        for (_i = 0, _len = _ref.length; _i < _len; _i++) {
          name = _ref[_i];
          _results.push(name.trim());
        }
        return _results;
      })();
    }
  }
  DashboardSwitcher.prototype.start = function(interval) {
    var pathParts;
    if (interval == null) {
      interval = 60000;
    }
    interval = parseInt(interval, 10);
    this.maxPos = this.dashboardNames.length - 1;
    if (this.dashboardNames.length === 0) {
      return;
    }
    pathParts = window.location.pathname.split('/');
    this.curName = pathParts[pathParts.length - 1];
    this.curPos = this.dashboardNames.indexOf(this.curName);
    if (this.curPos === -1) {
      this.curPos = 0;
      this.curName = this.dashboardNames[this.curPos];
    }
    this.switcherControls = new DashboardSwitcherControls(interval, this);
    if (this.switcherControls.present()) {
      this.switcherControls.start();
    }
    return this.startLoop(interval);
  };
  DashboardSwitcher.prototype.startLoop = function(interval) {
    var self;
    self = this;
    return this.handle = setTimeout(function() {
      self.curPos += 1;
      if (self.curPos > self.maxPos) {
        self.curPos = 0;
      }
      self.curName = self.dashboardNames[self.curPos];
      return window.location.pathname = "/" + self.curName;
    }, interval);
  };
  DashboardSwitcher.prototype.stopLoop = function() {
    return clearTimeout(this.handle);
  };
  DashboardSwitcher.prototype.currentName = function() {
    return this.curName;
  };
  DashboardSwitcher.prototype.nextName = function() {
    return this.dashboardNames[this.curPos + 1] || this.dashboardNames[0];
  };
  DashboardSwitcher.prototype.previousName = function() {
    return this.dashboardNames[this.curPos - 1] || this.dashboardNames[this.dashboardNames.length - 1];
  };
  return DashboardSwitcher;
})();
WidgetSwitcher = (function() {
  function WidgetSwitcher(elements) {
    this.elements = elements;
    this.$elements = $(this.elements);
  }
  WidgetSwitcher.prototype.start = function(interval) {
    var self;
    if (interval == null) {
      interval = 5000;
    }
    self = this;
    this.maxPos = this.$elements.length - 1;
    this.curPos = Math.min(1, this.maxPos);
    self.$elements.slice(1).hide();
    return this.handle = setInterval(function() {
      self.$elements.hide();
      $(self.$elements[self.curPos]).show().css('display', 'table-cell');
      self.curPos += 1;
      if (self.curPos > self.maxPos) {
        return self.curPos = 0;
      }
    }, parseInt(interval, 10));
  };
  WidgetSwitcher.prototype.stop = function() {
    return clearInterval(this.handle);
  };
  return WidgetSwitcher;
})();
DashboardSwitcherControls = (function() {
  var arrowContent, startTimerContent, stopTimerContent;
  arrowContent = "&#65515;";
  stopTimerContent = "stop timer";
  startTimerContent = "start timer";
  function DashboardSwitcherControls(interval, dashboardSwitcher) {
    if (interval == null) {
      interval = 60000;
    }
    this.updateTimer = __bind(this.updateTimer, this);
    this.isRunning = __bind(this.isRunning, this);
    this.pause = __bind(this.pause, this);
    this.pad = __bind(this.pad, this);
    this.currentTime = parseInt(interval, 10);
    this.interval = parseInt(interval, 10);
    this.$elements = $('#dc-switcher-controls');
    this.dashboardSwitcher = dashboardSwitcher;
    this.incrementTime = 1000;
    this.arrowContent = this.$elements.data('next-dashboard-content') || DashboardSwitcherControls.arrowContent;
    this.stopTimerContent = this.$elements.data('stop-timer-content') || DashboardSwitcherControls.stopTimerContent;
    this.startTimerContent = this.$elements.data('start-timer-content') || DashboardSwitcherControls.startTimerContent;
    this;
  }
  DashboardSwitcherControls.prototype.present = function() {
    return this.$elements.length;
  };
  DashboardSwitcherControls.prototype.start = function() {
    this.addElements();
    return this.$timer = $.timer(this.updateTimer, this.incrementTime, true);
  };
  DashboardSwitcherControls.prototype.addElements = function() {
    var template;
    template = this.$elements.find('dashboard-name-template');
    if (template.length) {
      this.$nextDashboardNameTemplate = template;
      this.$nextDashboardNameTemplate.remove();
    } else {
      this.$nextDashboardNameTemplate = $("<dashboard-name-template>Next dashboard: $nextName in </dashboard-name-template>");
    }
    this.$nextDashboardNameContainer = $("<span id='dc-switcher-next-name'></span>");
    this.$countdown = $("<span id='dc-switcher-countdown'></span>");
    this.$manualSwitcher = $("<span id='dc-switcher-next' class='fa fa-forward'></span>").html(this.arrowContent).click(__bind(function() {
      return location.href = "/" + (this.dashboardSwitcher.nextName());
    }, this));
    this.$switcherStopper = $("<span id='dc-switcher-pause-reset' class='fa fa-pause'></span>").html(this.stopTimerContent).click(this.pause);
    return this.$elements.append(this.$nextDashboardNameContainer).append(this.$countdown).append(this.$manualSwitcher).append(this.$switcherStopper);
  };
  DashboardSwitcherControls.prototype.formatTime = function(time) {
    var min, sec;
    time = time / 10;
    min = parseInt(time / 6000, 10);
    sec = parseInt(time / 100, 10) - (min * 60);
    return "" + (min > 0 ? this.pad(min, 2) : "00") + ":" + (this.pad(sec, 2));
  };
  DashboardSwitcherControls.prototype.pad = function(number, length) {
    var str;
    str = "" + number;
    while (str.length < length) {
      str = "0" + str;
    }
    return str;
  };
  DashboardSwitcherControls.prototype.pause = function() {
    this.$timer.toggle();
    if (this.isRunning()) {
      this.dashboardSwitcher.stopLoop();
      return this.$switcherStopper.removeClass('fa-pause').addClass('fa-play').html(this.startTimerContent);
    } else {
      this.dashboardSwitcher.startLoop(this.currentTime);
      return this.$switcherStopper.removeClass('fa-play').addClass('fa-pause').html(this.stopTimerContent);
    }
  };
  DashboardSwitcherControls.prototype.isRunning = function() {
    return this.$switcherStopper.hasClass('fa-pause');
  };
  DashboardSwitcherControls.prototype.resetCountdown = function() {
    var newTime;
    newTime = this.interval;
    if (newTime > 0) {
      this.currentTime = newTime;
    }
    return this.$timer.stop().once();
  };
  DashboardSwitcherControls.prototype.updateTimer = function() {
    var timeString;
    this.$nextDashboardNameContainer.html(this.$nextDashboardNameTemplate.html().replace('$nextName', this.dashboardSwitcher.nextName()));
    timeString = this.formatTime(this.currentTime);
    this.$countdown.html(timeString);
    if (this.currentTime === 0) {
      this.pause();
      this.resetCountdown();
      return;
    }
    this.currentTime -= this.incrementTime;
    if (this.currentTime < 0) {
      return this.currentTime = 0;
    }
  };
  return DashboardSwitcherControls;
})();
Dashing.DashboardSwitcher = DashboardSwitcher;
Dashing.WidgetSwitcher = WidgetSwitcher;
Dashing.DashboardSwitcherControls = DashboardSwitcherControls;
Dashing.on('ready', function() {
  var $container, ditcher;
  $('.gridster li').each(function(index, listItem) {
    var $listItem, $widgets, switcher;
    $listItem = $(listItem);
    $widgets = $listItem.children('div');
    if ($widgets.length > 1) {
      switcher = new WidgetSwitcher($widgets);
      return switcher.start($listItem.attr('data-switcher-interval') || 5000);
    }
  });
  $container = $('#container');
  ditcher = new DashboardSwitcher();
  return ditcher.start($container.attr('data-switcher-interval') || 60000);
});