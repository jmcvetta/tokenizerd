/*
                                   Gokenizer
                                  Test Suite

NOTE: Gokenizer application must be running in order to run tests.


@author: Jason McVetta <jason.mcvetta@gmail.com>
@copyright: (c) 2012 Jason McVetta
@license: GPL v3 - http://www.gnu.org/copyleft/gpl.html

********************************************************************************
This file is part of Gokenizer.

Gokenizer is free software: you can redistribute it and/or modify it under the
terms of the GNU General Public License as published by the Free Software
Foundation, either version 3 of the License, or (at your option) any later
version.

Gokenizer is distributed in the hope that it will be useful, but WITHOUT ANY
WARRANTY; without even the implied warranty of MERCHANTABILITY or FITNESS FOR A
PARTICULAR PURPOSE.  See the GNU General Public License for more details.

You should have received a copy of the GNU General Public License along with
Gokenizer.  If not, see <http://www.gnu.org/licenses/>.
********************************************************************************

*/

package tokenizer

import (
	"fmt"
	"github.com/jmcvetta/goutil"
	"launchpad.net/mgo"
	"log"
	"testing"
)

// Tests tokenization 
func TestRoundTrip(t *testing.T) {
	log.SetFlags(log.Ltime | log.Lshortfile)
	session, err := mgo.Dial("localhost")
	if err != nil {
		t.Error(err)
	}
	db := session.DB("test_tokenizer")
	err = db.DropDatabase()
	if err != nil {
		t.Error(err)
	}
	tokenizer := NewTokenizer(db)
	orig := goutil.RandAlphanumeric(8, 8)
	token := tokenizer.Tokenize(orig)
	var detok string // Result of detokenization - should be same as orig
	detok, err = tokenizer.Detokenize(token)
	if err != nil {
		t.Error(err)
	}
	if detok != orig {
		msg := "Detokenization failed: '%s' != '%s'."
		msg = fmt.Sprintf(msg, orig, detok)
		t.Error(msg)
	}
}

// Tests tokenization 
func BenchmarkRoundTrip(b *testing.B) {
	b.StopTimer()
	log.SetFlags(log.Ltime | log.Lshortfile)
	session, err := mgo.Dial("localhost")
	if err != nil {
		b.Error(err)
	}
	db := session.DB("test_tokenizer")
	err = db.DropDatabase()
	if err != nil {
		b.Error(err)
	}
	tokenizer := NewTokenizer(db)
	b.StartTimer()
	for i := 0; i < b.N; i++ {
		orig := goutil.RandAlphanumeric(8, 8)
		token := tokenizer.Tokenize(orig)
		var detok string // Result of detokenization - should be same as orig
		detok, err = tokenizer.Detokenize(token)
		if err != nil {
			b.Error(err)
		}
		if detok != orig {
			msg := "Detokenization failed: '%s' != '%s'."
			msg = fmt.Sprintf(msg, orig, detok)
			b.Error(msg)
		}
	}
}
