#ifndef LF_VI_STANDALONE_COMPAT_H
#define LF_VI_STANDALONE_COMPAT_H

#include <errno.h>
#include <stdio.h>
#include <stdlib.h>
#include <string.h>
#include <sys/ioctl.h>
#include <termios.h>
#include <unistd.h>

/* BusyBox's config-buffer helper was a stack buffer in this small use. */
#define RESERVE_CONFIG_BUFFER(name, size) char name[size]

static inline void *xmalloc(size_t size)
{
	void *p = malloc(size ? size : 1);
	if (p == NULL) {
		perror("vi: malloc");
		exit(EXIT_FAILURE);
	}
	return p;
}

static inline char *bb_xstrdup(const char *s)
{
	char *p = strdup(s ? s : "");
	if (p == NULL) {
		perror("vi: strdup");
		exit(EXIT_FAILURE);
	}
	return p;
}

static inline char *last_char_is(const char *s, int c)
{
	size_t len;
	if (s == NULL || *s == '\0') {
		return NULL;
	}
	len = strlen(s);
	return s[len - 1] == c ? (char *)(s + len - 1) : NULL;
}

static inline void get_terminal_width_height(int fd, int *width, int *height)
{
	struct winsize ws;
	if (ioctl(fd, TIOCGWINSZ, &ws) == 0) {
		if (ws.ws_col != 0) {
			*width = ws.ws_col;
		}
		if (ws.ws_row != 0) {
			*height = ws.ws_row;
		}
	}
}

#endif
