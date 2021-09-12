import { NgModule } from '@angular/core';
import { BrowserModule } from '@angular/platform-browser';

import { AppRoutingModule } from './app-routing.module';
import { AppComponent } from './app.component';
import { CoreModule } from './core/core.module';
import { PostLoginComponent } from './pages/postlogin/postlogin.component';
import { LoginComponent } from './pages/login/login.component';
import { SecretComponent } from './pages/secret/secret.component';
import { HttpClientModule } from '@angular/common/http';
import { FormsModule, ReactiveFormsModule } from '@angular/forms';
import { RouterModule } from '@angular/router';



@NgModule({
  declarations: [
    AppComponent,
    LoginComponent,
    PostLoginComponent,
    SecretComponent
  ],
  imports: [
    BrowserModule,
    AppRoutingModule,
    HttpClientModule,
    FormsModule,
    ReactiveFormsModule,
    RouterModule,

    CoreModule
  ],
  providers: [],
  bootstrap: [AppComponent]
})
export class AppModule { }
