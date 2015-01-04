'use strict';

var gulp = require('gulp'),
    uglify = require('gulp-uglify'),
    mainBowerFiles = require('main-bower-files'),
    gulpif = require('gulp-if'),
    gutil = require('gulp-util'),
    isRelease = gutil.env.release;

gulp.task('bower', function() {
  return gulp.src(mainBowerFiles())
    .pipe(gulpif(isRelease, uglify({preserveComments:'some'})))
    .pipe(gulp.dest('assets/javascripts/lib'));
});

gulp.task('default', ['bower']);
