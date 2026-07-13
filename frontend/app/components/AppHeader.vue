<script setup lang="ts"></script>

<template>
    <header class="header">
        App Header
        
        <nav class="header__nav">
            <ul class="header__list">
                <li class="header__item">
                    <NuxtLink to="/">
                        Home
                    </NuxtLink>
                </li>

                <li class="header__item">
                    <NuxtLink to="/users">
                        Users
                    </NuxtLink>
                </li>
            </ul>
        </nav>
    </header>
</template>

<style lang="scss" scoped></style>