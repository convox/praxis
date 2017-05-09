FROM convox/rails

COPY Gemfile Gemfile.lock ./
COPY vendor/cache ./vendor/cache
COPY .cache/build/gems /var/cache/build/gems
RUN bundle install --path /var/cache/build/gems

COPY Rakefile ./
COPY app/assets ./app/assets
COPY config ./config
COPY public ./public
RUN rake assets:precompile

COPY . .
