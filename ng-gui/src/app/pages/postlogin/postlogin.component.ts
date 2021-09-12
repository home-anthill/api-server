/*
 * MIT License
 *
 * Copyright (c) 2017-2019 Stefano Cappa
 *
 * Permission is hereby granted, free of charge, to any person obtaining a copy
 * of this software and associated documentation files (the "Software"), to deal
 * in the Software without restriction, including without limitation the rights
 * to use, copy, modify, merge, publish, distribute, sublicense, and/or sell
 * copies of the Software, and to permit persons to whom the Software is
 * furnished to do so, subject to the following conditions:
 *
 * The above copyright notice and this permission notice shall be included in all
 * copies or substantial portions of the Software.
 *
 * THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND, EXPRESS OR
 * IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF MERCHANTABILITY,
 * FITNESS FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT. IN NO EVENT SHALL THE
 * AUTHORS OR COPYRIGHT HOLDERS BE LIABLE FOR ANY CLAIM, DAMAGES OR OTHER
 * LIABILITY, WHETHER IN AN ACTION OF CONTRACT, TORT OR OTHERWISE, ARISING FROM,
 * OUT OF OR IN CONNECTION WITH THE SOFTWARE OR THE USE OR OTHER DEALINGS IN THE
 * SOFTWARE.
 */

import { Component, OnDestroy, OnInit } from '@angular/core';
import { ActivatedRoute, Params, Router } from '@angular/router';
import { Observable, Subscription } from 'rxjs';

import { AuthService } from '../../core/services/auth.service';
import { ExampleService } from '../../core/services/example.service';
import { tap } from 'rxjs/operators';

/**
 * Component to login
 */
@Component({
  selector: 'app-postlogin',
  templateUrl: 'postlogin.html'
})
export class PostLoginComponent implements OnInit, OnDestroy {
  private subscription: Subscription;

  constructor(private authService: AuthService, private router: Router, private activatedRoute: ActivatedRoute) {
    this.subscription = this.activatedRoute.queryParams
      .subscribe((params: Params) => {
        const token: string = params.token;
        this.authService.setToken(token);
        this.router.navigate(['/secret']);
      });
  }

  ngOnInit() {

  }

  ngOnDestroy() {
    // unsubscribe to all Subscriptions to prevent memory leaks and wrong behaviour
    if (this.subscription) {
      this.subscription.unsubscribe();
    }
  }
}
