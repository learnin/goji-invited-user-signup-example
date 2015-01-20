'use strict';

var gulp = require('gulp'),
    gutil = require('gulp-util'),
    sourcemaps = require('gulp-sourcemaps'),
    source = require('vinyl-source-stream'),
    buffer = require('vinyl-buffer'),
    watchify = require('watchify'),
    browserify = require('browserify'),
    uglify = require('gulp-uglify'),
    gulpif = require('gulp-if'),
    isRelease = !!gutil.env.release;

var bundler = browserify({
  entries: [
    './src/app.js',
    './bower_components/angular-bootstrap/ui-bootstrap-tpls.js'
  ],
  debug: !isRelease
}, watchify.args);

var bundle = function() {
  return bundler.bundle()
    .on('error', gutil.log.bind(gutil, 'Browserify Error'))
    .pipe(source('bundle.js'))
    .pipe(buffer())
    .pipe(sourcemaps.init({loadMaps: true}))
    .pipe(gulpif(isRelease, uglify({preserveComments:'some'})))
    .pipe(sourcemaps.write('./'))
    .pipe(gulp.dest('./assets/javascripts'));
};

if (!isRelease) {
  bundler = watchify(bundler);
  bundler.on('update', bundle);
}
bundler.transform('brfs');

gulp.task('browserify', bundle);

gulp.task('default', ['browserify']);
