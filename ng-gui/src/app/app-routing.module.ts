import { NgModule } from '@angular/core';
import { PreloadAllModules, RouterModule, Routes } from '@angular/router';
import { LoginComponent } from './pages/login/login.component';
import { PostLoginComponent } from './pages/postlogin/postlogin.component';
import { SecretComponent } from './pages/secret/secret.component';

const routes: Routes = [
  { path: '', component: LoginComponent },
  { path: 'postlogin', component: PostLoginComponent },
  { path: 'secret', component: SecretComponent },
];

@NgModule({
  imports: [RouterModule.forRoot(routes, {
    preloadingStrategy: PreloadAllModules
  })],
  exports: [RouterModule]
})
export class AppRoutingModule { }
